package balancer

import (
	"context"
	"errors"

	"github.com/fyerfyer/fyer-rpc/discovery/metrics"
	"github.com/fyerfyer/fyer-rpc/naming"
)

// BalancerType 负载均衡器类型
type BalancerType string

var (
	ErrNoAvailableInstances = errors.New("no available instances")
	ErrBalancerNotFound     = errors.New("balancer not found")
)

const (
	FastestResponse BalancerType = "fastest_response" // 最快响应时间
	Random          BalancerType = "random"           // 随机
	RoundRobin      BalancerType = "round_robin"      // 轮询
)

// Config 负载均衡器配置
type Config struct {
	Type           BalancerType    // 负载均衡类型
	MetricsClient  metrics.Metrics // 指标收集客户端
	UpdateInterval int64           // 更新间隔(秒)
	RetryTimes     int             // 重试次数
}

// Balancer 负载均衡器接口
type Balancer interface {
	// Initialize 初始化负载均衡器
	Initialize(instances []*naming.Instance) error

	// Select 选择一个服务实例
	Select(ctx context.Context) (*naming.Instance, error)

	// Update 更新服务实例列表
	Update(instances []*naming.Instance) error

	// Feedback 服务调用结果反馈，用于更新实例状态
	Feedback(ctx context.Context, instance *naming.Instance, duration int64, err error)

	// Name 返回负载均衡器名称
	Name() string
}

// Factory 负载均衡器工厂方法类型
type Factory func(conf *Config) Balancer

var factories = make(map[BalancerType]Factory)

// Register 注册负载均衡器工厂方法
func Register(typ BalancerType, factory Factory) {
	factories[typ] = factory
}

// Build 构建负载均衡器实例
func Build(conf *Config) (Balancer, error) {
	if factory, ok := factories[conf.Type]; ok {
		return factory(conf), nil
	}
	return nil, ErrBalancerNotFound
}
