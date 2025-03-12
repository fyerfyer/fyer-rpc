package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/fyerfyer/fyer-rpc/example/common"
	"github.com/fyerfyer/fyer-rpc/example/helloworld"
	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/fyerfyer/fyer-rpc/registry"
	_ "github.com/fyerfyer/fyer-rpc/registry/etcd"
)

// GreetServer 是 GreetService 的服务端实现
type GreetServer struct {
	greetService *helloworld.GreetServiceImpl // 原始服务实现
	registry     registry.Registry
	instance     *naming.Instance
	config       *common.ServerConfig
	detector     *HealthDetector
	metrics      *common.SimpleMetrics
	requestCount int64 // 请求计数
}

// NewGreetServer 创建新的 GreetServer 实例
func NewGreetServer(reg registry.Registry, config *common.ServerConfig) *GreetServer {
	addr := config.Address + ":" + strconv.Itoa(config.Port)

	// 创建指标收集器
	metrics := common.NewSimpleMetrics(100, 50)

	// 创建健康检测器
	detector := NewHealthDetector(config, metrics)

	server := &GreetServer{
		greetService: helloworld.NewGreetService(),
		registry:     reg,
		config:       config,
		detector:     detector,
		metrics:      metrics,
		instance: &naming.Instance{
			ID:      config.ID,
			Service: "GreetService",
			Version: "1.0.0",
			Address: addr,
			Status:  naming.StatusEnabled,
			Metadata: map[string]string{
				"protocol": "tcp",
				"weight":   "100",
				"serverID": config.ID,
			},
		},
	}

	return server
}

// SayHello 包装原始的 SayHello 方法，添加故障模拟逻辑
func (s *GreetServer) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloResponse, error) {
	// 增加请求计数
	reqCount := atomic.AddInt64(&s.requestCount, 1)

	// 检查是否应该模拟故障
	isHealthy := s.detector.IsHealthy()
	if !isHealthy {
		return nil, fmt.Errorf("service %s is currently unavailable", s.config.ID)
	}

	// 记录请求开始时间
	startTime := time.Now()

	// 调用原始方法
	resp, err := s.greetService.SayHello(ctx, req)

	// 记录指标
	duration := time.Since(startTime)
	s.metrics.RecordRequest(s.instance.Address, duration, err)

	log.Printf("Server %s handled request #%d for %s in %v",
		s.config.ID, reqCount, req.Name, duration)

	return resp, err
}

// GetGreetStats 包装原始的 GetGreetStats 方法
func (s *GreetServer) GetGreetStats(ctx context.Context, req *helloworld.StatsRequest) (*helloworld.StatsResponse, error) {
	return s.greetService.GetGreetStats(ctx, req)
}

// GreetServiceImpl 返回可以用于注册的服务实现
func (s *GreetServer) GreetServiceImpl() *helloworld.GreetServiceImpl {
	return s.greetService
}

// Start 启动服务并注册到注册中心
func (s *GreetServer) Start(ctx context.Context) error {
	// 注册服务实例
	err := s.registry.Register(ctx, s.instance)
	if err != nil {
		return fmt.Errorf("failed to register service: %v", err)
	}

	// 启动健康检测HTTP服务
	healthPort := strconv.Itoa(s.config.Port + 10000) // 健康检查端口：服务端口+10000
	s.detector.Start(":" + healthPort)
	log.Printf("Health detector for %s started on port %s", s.config.ID, healthPort)

	// 启动心跳
	go s.heartbeat(ctx)

	log.Printf("Server %s started at %s with failure config: rate=%.2f, after=%d, duration=%v",
		s.config.ID, s.instance.Address, s.config.FailRate, s.config.FailAfter, s.config.FailDuration)

	return nil
}

// Stop 停止服务并从注册中心注销
func (s *GreetServer) Stop(ctx context.Context) error {
	return s.registry.Deregister(ctx, s.instance)
}

// heartbeat 定期发送心跳
func (s *GreetServer) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("UpdateService for %s stopped", s.config.ID)
			return
		case <-ticker.C:
			err := s.registry.UpdateService(ctx, s.instance)
			if err != nil {
				log.Printf("UpdateService error for %s: %v", s.config.ID, err)
			}
		}
	}
}

// GetMetrics 获取服务指标
func (s *GreetServer) GetMetrics() *common.SimpleMetrics {
	return s.metrics
}

// GetStatus 获取服务状态
func (s *GreetServer) GetStatus() map[string]interface{} {
	isHealthy := s.detector.IsHealthy()
	status := "UP"
	if !isHealthy {
		status = "DOWN"
	}

	return map[string]interface{}{
		"id":           s.config.ID,
		"address":      s.instance.Address,
		"status":       status,
		"requestCount": atomic.LoadInt64(&s.requestCount),
		"failRate":     s.config.FailRate,
		"failAfter":    s.config.FailAfter,
		"failDuration": s.config.FailDuration,
		"timestamp":    time.Now().Unix(),
	}
}

// ResetCounter 重置请求计数
func (s *GreetServer) ResetCounter() {
	atomic.StoreInt64(&s.requestCount, 0)
	s.detector.Reset()
}

// GetHealthCheckURL 获取健康检查URL
func (s *GreetServer) GetHealthCheckURL() string {
	port := s.config.Port + 10000
	return fmt.Sprintf("http://%s:%d/health", s.config.Address, port)
}
