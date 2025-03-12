package discovery

import (
	"context"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/cluster/failover"
	"github.com/fyerfyer/fyer-rpc/discovery/balancer"
	"github.com/fyerfyer/fyer-rpc/discovery/metrics"
	"github.com/fyerfyer/fyer-rpc/naming"
)

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	resolver       *Resolver                        // 服务解析器
	balancer       balancer.Balancer                // 负载均衡算法实现
	failover       *failover.DefaultFailoverHandler // 故障转移处理器
	metrics        metrics.Metrics                  // 指标收集器
	serviceName    string                           // 服务名称
	version        string                           // 服务版本
	updateChan     <-chan struct{}                  // 服务更新通知通道
	closed         chan struct{}                    // 关闭信号
	enableFailover bool                             // 是否启用故障转移
	mu             sync.RWMutex
}

// LoadBalancerOption 负载均衡器配置选项
type LoadBalancerOption func(*LoadBalancer)

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(serviceName, version string, resolver *Resolver, metrics metrics.Metrics, balancerType balancer.BalancerType, options ...LoadBalancerOption) (*LoadBalancer, error) {
	// 创建负载均衡器
	b, err := balancer.Build(&balancer.Config{
		Type:           balancerType,
		MetricsClient:  metrics,
		UpdateInterval: 10, // 10秒更新一次
		RetryTimes:     3,  // 重试3次
	})
	if err != nil {
		return nil, err
	}

	// 获取初始服务列表
	instances, err := resolver.Resolve()
	if err != nil {
		return nil, err
	}

	// 初始化负载均衡器
	if err := b.Initialize(instances); err != nil {
		return nil, err
	}

	// 监听服务变更
	updateChan, err := resolver.Watch()
	if err != nil {
		return nil, err
	}

	lb := &LoadBalancer{
		resolver:       resolver,
		balancer:       b,
		metrics:        metrics,
		serviceName:    serviceName,
		version:        version,
		updateChan:     updateChan,
		closed:         make(chan struct{}),
		enableFailover: false, // 默认不启用故障转移
	}

	// 应用配置选项
	for _, option := range options {
		option(lb)
	}

	// 启动更新处理
	go lb.watchUpdates()

	return lb, nil
}

// WithFailover 启用故障转移功能
func WithFailover(failoverConfig *failover.Config) LoadBalancerOption {
	return func(lb *LoadBalancer) {
		handler, err := failover.NewFailoverHandler(failoverConfig)
		if err == nil {
			lb.failover = handler
			lb.enableFailover = true
		}
	}
}

// Select 选择一个服务实例
func (lb *LoadBalancer) Select(ctx context.Context) (*naming.Instance, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	return lb.balancer.Select(ctx)
}

// SelectWithFailover 选择一个服务实例并执行操作，支持故障转移
func (lb *LoadBalancer) SelectWithFailover(ctx context.Context, operation func(context.Context, *naming.Instance) error) error {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if !lb.enableFailover {
		// 不使用故障转移，直接选择实例并执行操作
		instance, err := lb.balancer.Select(ctx)
		if err != nil {
			return err
		}

		startTime := time.Now()
		err = operation(ctx, instance)
		duration := time.Since(startTime)

		// 反馈结果
		lb.Feedback(ctx, instance, duration.Milliseconds(), err)
		return err
	}

	// 获取当前可用实例
	instances, err := lb.resolver.Resolve()
	if err != nil {
		return err
	}

	// 使用故障转移处理器执行操作
	result, err := lb.failover.Execute(ctx, instances, operation)
	if err != nil {
		return err
	}

	// 如果操作成功，更新负载均衡器中的实例状态
	if result.Success && result.Instance != nil {
		lb.balancer.Feedback(ctx, result.Instance, result.Duration.Milliseconds(), nil)
	}

	return nil
}

// Feedback 反馈调用结果
func (lb *LoadBalancer) Feedback(ctx context.Context, instance *naming.Instance, duration int64, err error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// 更新负载均衡器状态
	lb.balancer.Feedback(ctx, instance, duration, err)

	// 如果启用了故障转移，也更新故障转移组件的状态
	if lb.enableFailover && instance != nil {
		if err != nil {
			lb.failover.GetDetector().MarkFailed(ctx, instance)
		} else {
			lb.failover.GetDetector().MarkSuccess(ctx, instance)
		}
	}
}

// watchUpdates 监听服务更新
func (lb *LoadBalancer) watchUpdates() {
	for {
		select {
		case <-lb.closed:
			return
		case <-lb.updateChan:
			// 获取最新的服务列表
			instances, err := lb.resolver.Resolve()
			if err != nil {
				continue
			}

			// 更新负载均衡器
			lb.mu.Lock()
			err = lb.balancer.Update(instances)
			lb.mu.Unlock()

			if err != nil {
				// 可以添加日志记录
				continue
			}
		}
	}
}

// UpdateInstances 手动更新服务实例列表
func (lb *LoadBalancer) UpdateInstances(instances []*naming.Instance) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	return lb.balancer.Update(instances)
}

// GetInstances 获取当前可用的服务实例列表
func (lb *LoadBalancer) GetInstances() ([]*naming.Instance, error) {
	return lb.resolver.Resolve()
}

// Close 关闭负载均衡器
func (lb *LoadBalancer) Close() error {
	close(lb.closed)
	return lb.resolver.Close()
}

// Stats 获取负载均衡统计信息
type Stats struct {
	ServiceName    string
	Version        string
	TotalInstances int
	HealthyCount   int
	Latencies      map[string]time.Duration
}

// GetStats 获取负载均衡器统计信息
func (lb *LoadBalancer) GetStats(ctx context.Context) (*Stats, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// 获取当前服务实例列表
	instances, err := lb.resolver.Resolve()
	if err != nil {
		return nil, err
	}

	// 获取所有实例的延迟信息
	latencies, err := lb.metrics.GetServiceLatency(ctx, lb.serviceName)
	if err != nil {
		return nil, err
	}

	// 统计健康实例数量
	healthyCount := 0
	for _, instance := range instances {
		if instance.Status == naming.StatusEnabled {
			healthyCount++
		}
	}

	return &Stats{
		ServiceName:    lb.serviceName,
		Version:        lb.version,
		TotalInstances: len(instances),
		HealthyCount:   healthyCount,
		Latencies:      latencies,
	}, nil
}

// IsFailoverEnabled 检查是否启用了故障转移功能
func (lb *LoadBalancer) IsFailoverEnabled() bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.enableFailover
}

// GetFailoverHandler 获取故障转移处理器
func (lb *LoadBalancer) GetFailoverHandler() *failover.DefaultFailoverHandler {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.failover
}
