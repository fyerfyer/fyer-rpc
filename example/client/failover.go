package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/cluster/failover"
	"github.com/fyerfyer/fyer-rpc/example/common"
	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/fyerfyer/fyer-rpc/protocol/codec"
	"github.com/fyerfyer/fyer-rpc/rpc"
)

// FailoverManager 管理客户端的故障转移功能
type FailoverManager struct {
	handler        *failover.DefaultFailoverHandler // 故障转移处理器
	config         *failover.Config                 // 故障转移配置
	metrics        *common.SimpleMetrics            // 指标收集器
	serverList     []*naming.Instance               // 服务器实例列表
	activeInstance *naming.Instance                 // 当前活跃的实例
	mu             sync.RWMutex
}

// NewFailoverManager 创建一个新的故障转移管理器
func NewFailoverManager(config *failover.Config, serverAddresses []string) (*FailoverManager, error) {
	// 创建故障转移处理器
	handler, err := failover.NewFailoverHandler(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create failover handler: %w", err)
	}

	// 创建指标收集器
	metrics := common.NewSimpleMetrics(100, 30)

	// 将地址转换为实例
	instances := make([]*naming.Instance, len(serverAddresses))
	for i, addr := range serverAddresses {
		instances[i] = &naming.Instance{
			ID:      fmt.Sprintf("server-%d", i),
			Service: "GreetService",
			Address: addr,
			Status:  naming.StatusEnabled,
		}
	}

	return &FailoverManager{
		handler:    handler,
		config:     config,
		metrics:    metrics,
		serverList: instances,
	}, nil
}

// ExecuteRPC 执行带故障转移的RPC调用
func (fm *FailoverManager) ExecuteRPC(ctx context.Context, serviceName, methodName string, req interface{}, resp interface{}) error {
	// 复制实例列表以防并发修改
	fm.mu.RLock()
	instances := make([]*naming.Instance, len(fm.serverList))
	copy(instances, fm.serverList)
	fm.mu.RUnlock()

	if len(instances) == 0 {
		return fmt.Errorf("no server instances available")
	}

	// 记录操作开始时间
	startTime := time.Now()

	// 定义RPC调用操作
	operation := func(ctx context.Context, instance *naming.Instance) error {
		// 创建到具体实例的新连接
		client, err := rpc.NewClient(instance.Address)
		if err != nil {
			fm.metrics.RecordRequest(instance.Address, time.Since(startTime), err)
			return err
		}
		defer client.Close()

		// 执行RPC调用
		data, err := client.CallWithTimeout(ctx, serviceName, methodName, req)
		if err != nil {
			fm.metrics.RecordRequest(instance.Address, time.Since(startTime), err)
			return err
		}

		// 获取codec并解码响应
		jsonCodec := codec.GetCodec(codec.JSON)
		if err := jsonCodec.Decode(data, resp); err != nil {
			fm.metrics.RecordRequest(instance.Address, time.Since(startTime), err)
			return err
		}

		// 记录成功指标
		fm.metrics.RecordRequest(instance.Address, time.Since(startTime), nil)

		// 更新活跃实例
		fm.mu.Lock()
		fm.activeInstance = instance
		fm.mu.Unlock()

		return nil
	}

	// 执行带故障转移的调用
	result, err := fm.handler.Execute(ctx, instances, operation)
	if err != nil {
		// 记录故障事件
		if result != nil && result.Instance != nil && result.RetryCount > 0 {
			log.Printf("Failover: %d retries performed, failed nodes: %v",
				result.RetryCount, result.FailedNodes)
			fm.metrics.RecordRetry(result.Instance.Address)
		}
		return err
	}

	// 记录成功的故障转移
	if result != nil && result.RetryCount > 0 {
		// 找出从哪个实例故障转移的
		var fromInstance string
		if len(result.FailedNodes) > 0 {
			fromInstance = result.FailedNodes[0]
		} else {
			fromInstance = "unknown"
		}

		fm.metrics.RecordFailover(fromInstance, result.Instance.Address)
		log.Printf("Successful failover from %s to %s after %d retries",
			fromInstance, result.Instance.Address, result.RetryCount)
	}

	return nil
}

// ReportCircuitBreak 报告熔断事件
func (fm *FailoverManager) ReportCircuitBreak(instanceAddr string, state failover.State) {
	stateStr := "unknown"
	switch state {
	case failover.StateClosed:
		stateStr = "closed"
	case failover.StateOpen:
		stateStr = "open"
	case failover.StateHalfOpen:
		stateStr = "half-open"
	}

	fm.metrics.RecordCircuitBreak(instanceAddr, stateStr)
}

// GetActiveInstance 获取当前活跃的实例
func (fm *FailoverManager) GetActiveInstance() *naming.Instance {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.activeInstance
}

// UpdateServerList 更新服务器实例列表
func (fm *FailoverManager) UpdateServerList(instances []*naming.Instance) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.serverList = instances
}

// AddServer 添加服务器实例
func (fm *FailoverManager) AddServer(address string) {
	instance := &naming.Instance{
		ID:      fmt.Sprintf("server-%d", len(fm.serverList)),
		Service: "GreetService",
		Address: address,
		Status:  naming.StatusEnabled,
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.serverList = append(fm.serverList, instance)
}

// RemoveServer 移除服务器实例
func (fm *FailoverManager) RemoveServer(address string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for i, instance := range fm.serverList {
		if instance.Address == address {
			// 从列表中删除此实例
			fm.serverList = append(fm.serverList[:i], fm.serverList[i+1:]...)
			return
		}
	}
}

// GetMetrics 获取指标数据
func (fm *FailoverManager) GetMetrics() *common.SimpleMetrics {
	return fm.metrics
}

// SetHealthCheckFunction 设置健康检查函数
func (fm *FailoverManager) SetHealthCheckFunction() {
	detector := fm.handler.GetDetector()
	if _, ok := detector.(*failover.TimeoutDetector); ok {
		// TimeoutDetector已经有实现，这里不需要特别设置
		log.Println("Using built-in TCP timeout detector for health checks")
	}
}

// IsInstanceHealthy 检查实例是否健康
func (fm *FailoverManager) IsInstanceHealthy(ctx context.Context, instance *naming.Instance) bool {
	detector := fm.handler.GetDetector()
	status, err := detector.Detect(ctx, instance)
	if err != nil {
		return false
	}
	return status == failover.StatusHealthy
}

// IsBreakerOpen 检查熔断器是否开启
func (fm *FailoverManager) IsBreakerOpen(ctx context.Context, instance *naming.Instance) bool {
	breaker := fm.handler.GetCircuitBreaker()
	allowed, _ := breaker.Allow(ctx, instance)
	return !allowed
}

// CheckAllServers 检查所有服务器的健康状态
func (fm *FailoverManager) CheckAllServers(ctx context.Context) map[string]bool {
	fm.mu.RLock()
	instances := make([]*naming.Instance, len(fm.serverList))
	copy(instances, fm.serverList)
	fm.mu.RUnlock()

	results := make(map[string]bool, len(instances))
	for _, instance := range instances {
		results[instance.Address] = fm.IsInstanceHealthy(ctx, instance)
	}

	return results
}

// TestConnection 测试到特定地址的连接是否正常
func (fm *FailoverManager) TestConnection(address string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
