package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/fyerfyer/fyer-rpc/discovery"
	"github.com/fyerfyer/fyer-rpc/discovery/balancer"
	"github.com/fyerfyer/fyer-rpc/discovery/metrics"
	"github.com/fyerfyer/fyer-rpc/example/common"
	"github.com/fyerfyer/fyer-rpc/example/helloworld"
	"github.com/fyerfyer/fyer-rpc/protocol/codec"
	"github.com/fyerfyer/fyer-rpc/registry/etcd"
	"github.com/fyerfyer/fyer-rpc/rpc"
)

// GreetClient 包装了问候服务客户端的实现
type GreetClient struct {
	balancer        *discovery.LoadBalancer
	metrics         metrics.Metrics
	discovery       discovery.Discovery
	failoverManager *FailoverManager     // 故障转移管理器
	config          *common.ClientConfig // 客户端配置
}

// NewGreetClient 创建问候服务客户端
func NewGreetClient(serviceName string, config *common.ClientConfig) (*GreetClient, error) {
	// 创建 etcd 注册中心客户端
	registry, err := etcd.New(
		etcd.WithEndpoints([]string{"localhost:2379"}),
		etcd.WithDialTimeout(time.Second*5),
	)
	if err != nil {
		return nil, err
	}

	// 创建服务发现
	disc := discovery.NewDiscovery(registry, time.Second*10)

	// 创建 Prometheus 指标收集器
	var metricsCollector metrics.Metrics
	prometheusMetrics, err := metrics.NewPrometheusMetrics(&metrics.PrometheusConfig{
		QueryURL: "http://localhost:9090",
		JobName:  "fyerrpc_client",
	})
	if err != nil {
		// 使用空操作指标收集器作为回退方案
		log.Printf("Warning: Failed to create Prometheus metrics collector: %v, using NoopMetrics instead", err)
		metricsCollector = metrics.NewNoopMetrics()
	} else {
		metricsCollector = prometheusMetrics
	}

	// 创建服务解析器
	resolver, err := discovery.NewResolver(registry, serviceName, "1.0.0")
	if err != nil {
		return nil, err
	}

	// 创建负载均衡器
	lb, err := discovery.NewLoadBalancer(serviceName, "1.0.0", resolver, metricsCollector, balancer.FastestResponse)
	if err != nil {
		return nil, err
	}

	client := &GreetClient{
		balancer:  lb,
		metrics:   metricsCollector,
		discovery: disc,
		config:    config,
	}

	// 如果启用了故障转移，创建故障转移管理器
	if config != nil && config.EnableFailover && config.FailoverConfig != nil {
		failoverManager, err := NewFailoverManager(config.FailoverConfig, config.ServerAddresses)
		if err != nil {
			log.Printf("Warning: Failed to create failover manager: %v", err)
		} else {
			client.failoverManager = failoverManager
			log.Printf("Failover enabled with %d servers", len(config.ServerAddresses))
		}
	}

	return client, nil
}

// SayHello 调用问候服务
func (c *GreetClient) SayHello(ctx context.Context, name string, greeting string) (*helloworld.HelloResponse, error) {
	// 构造请求
	req := &helloworld.HelloRequest{
		Name:     name,
		Greeting: greeting,
	}

	// 记录开始时间
	start := time.Now()

	// 准备响应对象
	resp := new(helloworld.HelloResponse)

	// 根据是否启用故障转移选择调用方式
	var err error
	if c.failoverManager != nil && c.config.EnableFailover {
		// 使用故障转移机制调用
		err = c.callWithFailover(ctx, "GreetService", "SayHello", req, resp)
	} else {
		// 使用负载均衡调用
		err = c.callWithLoadBalancer(ctx, "GreetService", "SayHello", req, resp)
	}

	// 记录调用结果
	if err != nil {
		log.Printf("SayHello failed: %v", err)
	} else {
		log.Printf("SayHello succeeded: %s", resp.Message)
	}

	duration := time.Since(start)
	log.Printf("SayHello took %v", duration)

	return resp, err
}

// GetStats 获取服务统计信息
func (c *GreetClient) GetStats(ctx context.Context) (*helloworld.StatsResponse, error) {
	req := &helloworld.StatsRequest{}
	resp := new(helloworld.StatsResponse)

	// 根据是否启用故障转移选择调用方式
	var err error
	if c.failoverManager != nil && c.config.EnableFailover {
		// 使用故障转移机制调用
		err = c.callWithFailover(ctx, "GreetService", "GetGreetStats", req, resp)
	} else {
		// 使用负载均衡调用
		err = c.callWithLoadBalancer(ctx, "GreetService", "GetGreetStats", req, resp)
	}

	return resp, err
}

// callWithLoadBalancer 使用负载均衡调用服务
func (c *GreetClient) callWithLoadBalancer(ctx context.Context, serviceName, methodName string, req interface{}, resp interface{}) error {
	// 使用负载均衡器选择服务实例
	instance, err := c.balancer.Select(ctx)
	if err != nil {
		return err
	}

	// 创建到该实例的客户端连接
	client, err := rpc.NewClient(instance.Address)
	if err != nil {
		c.balancer.Feedback(ctx, instance, 0, err)
		return err
	}
	defer client.Close()

	// 记录开始时间
	start := time.Now()

	// 发起调用
	data, err := client.CallWithTimeout(ctx, serviceName, methodName, req)
	if err != nil {
		c.balancer.Feedback(ctx, instance, time.Since(start).Nanoseconds(), err)
		return err
	}

	// 解码响应
	jsonCodec := codec.GetCodec(codec.JSON)
	if err := jsonCodec.Decode(data, resp); err != nil {
		c.balancer.Feedback(ctx, instance, time.Since(start).Nanoseconds(), err)
		return err
	}

	// 记录成功的调用
	c.balancer.Feedback(ctx, instance, time.Since(start).Nanoseconds(), nil)

	return nil
}

// callWithFailover 使用故障转移机制调用服务
func (c *GreetClient) callWithFailover(ctx context.Context, serviceName, methodName string, req interface{}, resp interface{}) error {
	if c.failoverManager == nil {
		return errors.New("failover manager is not initialized")
	}

	// 使用故障转移管理器执行RPC调用
	err := c.failoverManager.ExecuteRPC(ctx, serviceName, methodName, req, resp)
	if err != nil {
		// 如果有活跃实例，记录故障
		if instance := c.failoverManager.GetActiveInstance(); instance != nil {
			log.Printf("Failover: call to %s failed: %v", instance.Address, err)
		}
		return err
	}

	return nil
}

// Close 关闭客户端
func (c *GreetClient) Close() error {
	// 关闭相关资源
	if c.discovery != nil {
		c.discovery.Close()
	}
	if c.metrics != nil {
		c.metrics.Close()
	}
	return nil
}

// GetFailoverMetrics 获取故障转移指标
func (c *GreetClient) GetFailoverMetrics() *common.SimpleMetrics {
	if c.failoverManager != nil {
		return c.failoverManager.GetMetrics()
	}
	return nil
}
