package main

import (
	"context"
	"time"

	"github.com/fyerfyer/fyer-rpc/discovery"
	"github.com/fyerfyer/fyer-rpc/discovery/balancer"
	"github.com/fyerfyer/fyer-rpc/discovery/metrics"
	"github.com/fyerfyer/fyer-rpc/example/helloworld"
	"github.com/fyerfyer/fyer-rpc/protocol"
	"github.com/fyerfyer/fyer-rpc/registry/etcd"
	"github.com/fyerfyer/fyer-rpc/rpc"
)

// GreetClient 包装了问候服务客户端的实现
type GreetClient struct {
	client    *rpc.Client
	balancer  *discovery.LoadBalancer
	metrics   metrics.Metrics
	discovery discovery.Discovery
}

// NewGreetClient 创建问候服务客户端
func NewGreetClient(serviceName string) (*GreetClient, error) {
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
	metricsCollector, err := metrics.NewPrometheusMetrics(&metrics.PrometheusConfig{
		QueryURL:       "http://localhost:9090",
		JobName:        "fyerrpc_client",
	})
	if err != nil {
		return nil, err
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

	return &GreetClient{
		balancer:  lb,
		metrics:   metricsCollector,
		discovery: disc,
	}, nil
}

// SayHello 调用问候服务
func (c *GreetClient) SayHello(ctx context.Context, name string, greeting string) (*helloworld.HelloResponse, error) {
	// 使用负载均衡器选择服务实例
	instance, err := c.balancer.Select(ctx)
	if err != nil {
		return nil, err
	}

	// 创建或获取到该实例的客户端连接
	client, err := rpc.NewClient(instance.Address)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// 构造请求
	req := &helloworld.HelloRequest{
		Name:     name,
		Greeting: greeting,
	}

	// 记录开始时间
	start := time.Now()

	// 发起调用
	resp := new(helloworld.HelloResponse)
	// 修复：改用 CallWithTimeout，并传入正确的参数
	data, err := client.CallWithTimeout(ctx, "GreetService", "SayHello", req)
	if err != nil {
		// 记录失败的调用
		c.balancer.Feedback(ctx, instance, time.Since(start).Nanoseconds(), err)
		return nil, err
	}

	// 解码响应
	if err := protocol.GetCodecByType(protocol.SerializationTypeJSON).Decode(data, resp); err != nil {
		return nil, err
	}

	// 记录成功的调用
	c.balancer.Feedback(ctx, instance, time.Since(start).Nanoseconds(), nil)

	return resp, nil
}

// GetStats 获取服务统计信息
func (c *GreetClient) GetStats(ctx context.Context) (*helloworld.StatsResponse, error) {
	instance, err := c.balancer.Select(ctx)
	if err != nil {
		return nil, err
	}

	client, err := rpc.NewClient(instance.Address)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	req := &helloworld.StatsRequest{}
	resp := new(helloworld.StatsResponse)

	start := time.Now()
	// 修复：改用 CallWithTimeout，并传入正确的参数
	data, err := client.CallWithTimeout(ctx, "GreetService", "GetGreetStats", req)
	if err != nil {
		c.balancer.Feedback(ctx, instance, time.Since(start).Nanoseconds(), err)
		return nil, err
	}

	// 解码响应
	if err := protocol.GetCodecByType(protocol.SerializationTypeJSON).Decode(data, resp); err != nil {
		return nil, err
	}

	c.balancer.Feedback(ctx, instance, time.Since(start).Nanoseconds(), nil)
	return resp, nil
}

// Close 关闭客户端
func (c *GreetClient) Close() error {
	if c.balancer != nil {
		c.balancer.Close()
	}
	if c.metrics != nil {
		c.metrics.Close()
	}
	if c.discovery != nil {
		c.discovery.Close()
	}
	return nil
}
