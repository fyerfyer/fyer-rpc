package balancer

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/discovery/metrics"
	"github.com/fyerfyer/fyer-rpc/naming"
)

func init() {
	Register(FastestResponse, NewFastestBalancer)
}

// FastestBalancer 最快响应时间负载均衡器
type FastestBalancer struct {
	instances    []*instanceWrapper // 服务实例包装列表
	metrics      metrics.Metrics    // 指标收集客户端
	updateTicker *time.Ticker       // 更新定时器
	retryTimes   int                // 重试次数
	mu           sync.RWMutex
}

// instanceWrapper 实例包装，包含实例信息和性能指标
type instanceWrapper struct {
	*naming.Instance               // 服务实例
	latency          time.Duration // 平均响应时间
	weight           float64       // 权重分数
	lastUpdate       time.Time     // 最后更新时间
}

// NewFastestBalancer 创建最快响应时间负载均衡器
func NewFastestBalancer(conf *Config) Balancer {
	fb := &FastestBalancer{
		metrics:    conf.MetricsClient,
		retryTimes: conf.RetryTimes,
	}

	// 启动定时更新协程
	if conf.UpdateInterval > 0 {
		fb.updateTicker = time.NewTicker(time.Duration(conf.UpdateInterval) * time.Second)
		go fb.updateLoop()
	}

	return fb
}

// Initialize 初始化负载均衡器
func (b *FastestBalancer) Initialize(instances []*naming.Instance) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.updateInstances(instances)
}

// Select 选择一个服务实例
func (b *FastestBalancer) Select(ctx context.Context) (*naming.Instance, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.instances) == 0 {
		return nil, ErrNoAvailableInstances
	}

	// 按响应时间排序
	instances := make([]*instanceWrapper, len(b.instances))
	copy(instances, b.instances)
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].latency < instances[j].latency
	})

	// 选择响应最快的实例
	// 如果第一个实例调用失败，会重试其他实例
	for i := 0; i < b.retryTimes && i < len(instances); i++ {
		if instances[i].Status == naming.StatusEnabled {
			return instances[i].Instance, nil
		}
	}

	return nil, ErrNoAvailableInstances
}

// Update 更新服务实例列表
func (b *FastestBalancer) Update(instances []*naming.Instance) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.updateInstances(instances)
}

// Feedback 服务调用结果反馈
func (b *FastestBalancer) Feedback(ctx context.Context, instance *naming.Instance, duration int64, err error) {
	if err != nil {
		return // 出错时不更新统计信息
	}

	b.metrics.RecordResponse(ctx, &metrics.ResponseMetric{
		ServiceName: instance.Service,
		Instance:    instance.Address,
		Duration:    time.Duration(duration),
		Status:      "success",
		Timestamp:   time.Now(),
	})
}

// Name 返回负载均衡器名称
func (b *FastestBalancer) Name() string {
	return string(FastestResponse)
}

// updateInstances 更新实例列表并获取最新的性能指标
func (b *FastestBalancer) updateInstances(instances []*naming.Instance) error {
	// 创建新的实例包装列表
	newInstances := make([]*instanceWrapper, 0, len(instances))

	// 获取所有实例的性能指标
	for _, ins := range instances {
		latency, err := b.metrics.GetLatency(context.Background(), ins.Service, ins.Address)
		if err != nil {
			latency = time.Second // 默认延迟为1秒
		}

		wrapper := &instanceWrapper{
			Instance:   ins,
			latency:    latency,
			lastUpdate: time.Now(),
		}

		// 计算权重分数
		wrapper.weight = calculateWeight(latency)
		newInstances = append(newInstances, wrapper)
	}

	// 更新实例列表
	b.instances = newInstances
	return nil
}

// updateLoop 定时更新性能指标
func (b *FastestBalancer) updateLoop() {
	for range b.updateTicker.C {
		b.mu.RLock()
		instances := make([]*naming.Instance, len(b.instances))
		for i, wrapper := range b.instances {
			instances[i] = wrapper.Instance
		}
		b.mu.RUnlock()

		// 更新实例列表
		_ = b.updateInstances(instances)
	}
}

// calculateWeight 计算实例权重
// 使用响应时间的倒数作为权重，响应越快权重越大
func calculateWeight(latency time.Duration) float64 {
	if latency <= 0 {
		return 1.0
	}
	return 1.0 / float64(latency.Milliseconds())
}
