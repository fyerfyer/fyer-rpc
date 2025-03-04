package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/model"
)

// PrometheusMetrics Prometheus指标收集器
type PrometheusMetrics struct {
	// Prometheus客户端API
	api    v1.API
	pusher *push.Pusher

	// 指标定义
	responseTime  *prometheus.HistogramVec
	requestTotal  *prometheus.CounterVec
	failoverTotal *prometheus.CounterVec // 新增：故障转移计数
	circuitBreaks *prometheus.CounterVec // 新增：熔断事件计数
	retryTotal    *prometheus.CounterVec // 新增：重试计数
	failoverRate  *prometheus.GaugeVec   // 新增：故障转移率

	// 配置项
	pushGatewayURL string
	queryURL       string
	jobName        string

	mu sync.RWMutex
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	PushGatewayURL string        // Push gateway地址
	QueryURL       string        // Prometheus查询地址
	JobName        string        // 任务名称
	PushInterval   time.Duration // 推送间隔
}

// NewPrometheusMetrics 创建Prometheus指标收集器
func NewPrometheusMetrics(config *PrometheusConfig) (*PrometheusMetrics, error) {
	// 创建Prometheus客户端
	client, err := api.NewClient(api.Config{
		Address: config.QueryURL,
	})
	if err != nil {
		return nil, err
	}

	// 创建指标收集器
	pm := &PrometheusMetrics{
		api:            v1.NewAPI(client),
		pushGatewayURL: config.PushGatewayURL,
		queryURL:       config.QueryURL,
		jobName:        config.JobName,

		// 定义响应时间直方图
		responseTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rpc_response_time_seconds",
				Help:    "RPC response time in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "instance", "status"},
		),

		// 定义请求总数计数器
		requestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rpc_requests_total",
				Help: "Total number of RPC requests",
			},
			[]string{"service", "method", "instance", "status"},
		),

		// 新增：故障转移计数器
		failoverTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rpc_failover_total",
				Help: "Total number of failover events",
			},
			[]string{"service", "from_instance", "to_instance"},
		),

		// 新增：熔断器事件计数器
		circuitBreaks: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rpc_circuit_breaks_total",
				Help: "Total number of circuit breaker state changes",
			},
			[]string{"service", "instance", "state"},
		),

		// 新增：重试计数器
		retryTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rpc_retry_total",
				Help: "Total number of retry attempts",
			},
			[]string{"service", "instance", "attempt"},
		),

		// 新增：故障转移率
		failoverRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "rpc_failover_rate",
				Help: "Rate of failovers per service",
			},
			[]string{"service"},
		),
	}

	// 注册指标
	prometheus.MustRegister(pm.responseTime)
	prometheus.MustRegister(pm.requestTotal)
	prometheus.MustRegister(pm.failoverTotal)
	prometheus.MustRegister(pm.circuitBreaks)
	prometheus.MustRegister(pm.retryTotal)
	prometheus.MustRegister(pm.failoverRate)

	// 创建推送器
	pm.pusher = push.New(config.PushGatewayURL, config.JobName).
		Collector(pm.responseTime).
		Collector(pm.requestTotal).
		Collector(pm.failoverTotal).
		Collector(pm.circuitBreaks).
		Collector(pm.retryTotal).
		Collector(pm.failoverRate)

	// 启动定时推送
	if config.PushInterval > 0 {
		go pm.startPushing(config.PushInterval)
	}

	return pm, nil
}

// RecordResponse 记录响应时间
func (p *PrometheusMetrics) RecordResponse(ctx context.Context, metric *ResponseMetric) error {
	p.responseTime.WithLabelValues(
		metric.ServiceName,
		metric.MethodName,
		metric.Instance,
		metric.Status,
	).Observe(metric.Duration.Seconds())

	p.requestTotal.WithLabelValues(
		metric.ServiceName,
		metric.MethodName,
		metric.Instance,
		metric.Status,
	).Inc()

	return nil
}

// GetLatency 获取指定服务实例的平均响应时间
func (p *PrometheusMetrics) GetLatency(ctx context.Context, service, instance string) (time.Duration, error) {
	query := fmt.Sprintf(
		`rate(rpc_response_time_seconds_sum{service="%s",instance="%s"}[1m]) / 
         rate(rpc_response_time_seconds_count{service="%s",instance="%s"}[1m])`,
		service, instance, service, instance,
	)

	result, warnings, err := p.api.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	if len(warnings) > 0 {
		// 记录警告但继续处理
	}

	// 正确处理查询结果
	vector, ok := result.(model.Vector)
	if !ok || len(vector) == 0 {
		return time.Second, nil // 默认返回1秒
	}

	// 获取平均响应时间（秒）并转换为Duration
	latency := time.Duration(float64(vector[0].Value) * float64(time.Second))
	return latency, nil
}

// GetServiceLatency 获取服务所有实例的平均响应时间
func (p *PrometheusMetrics) GetServiceLatency(ctx context.Context, service string) (map[string]time.Duration, error) {
	query := fmt.Sprintf(
		`rate(rpc_response_time_seconds_sum{service="%s"}[1m]) / 
         rate(rpc_response_time_seconds_count{service="%s"}[1m])`,
		service, service,
	)

	result, warnings, err := p.api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		// 记录警告但继续处理
	}

	// 处理查询结果
	vector, ok := result.(model.Vector)
	if !ok {
		return make(map[string]time.Duration), nil
	}

	latencies := make(map[string]time.Duration)
	for _, sample := range vector {
		instance := string(sample.Metric["instance"])
		latency := time.Duration(float64(sample.Value) * float64(time.Second))
		latencies[instance] = latency
	}

	return latencies, nil
}

// RecordFailover 记录故障转移事件
func (p *PrometheusMetrics) RecordFailover(ctx context.Context, service, fromInstance, toInstance string) error {
	p.failoverTotal.WithLabelValues(
		service,
		fromInstance,
		toInstance,
	).Inc()

	// 更新故障转移率
	p.updateFailoverRate(ctx, service)

	return nil
}

// RecordCircuitBreak 记录熔断事件
func (p *PrometheusMetrics) RecordCircuitBreak(ctx context.Context, service, instance string, state string) error {
	p.circuitBreaks.WithLabelValues(
		service,
		instance,
		state,
	).Inc()

	return nil
}

// RecordRetry 记录重试事件
func (p *PrometheusMetrics) RecordRetry(ctx context.Context, service, instance string, attempt int) error {
	p.retryTotal.WithLabelValues(
		service,
		instance,
		fmt.Sprintf("%d", attempt),
	).Inc()

	return nil
}

// GetFailoverRate 获取故障转移率
func (p *PrometheusMetrics) GetFailoverRate(ctx context.Context, service string) (float64, error) {
	query := fmt.Sprintf(
		`sum(rate(rpc_failover_total{service="%s"}[5m])) / 
         sum(rate(rpc_requests_total{service="%s"}[5m]))`,
		service, service,
	)

	result, warnings, err := p.api.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	if len(warnings) > 0 {
		// 记录警告但继续处理
	}

	// 处理查询结果
	vector, ok := result.(model.Vector)
	if !ok || len(vector) == 0 {
		return 0, nil
	}

	// 获取故障转移率
	rate := float64(vector[0].Value)
	return rate, nil
}

// updateFailoverRate 更新故障转移率指标
func (p *PrometheusMetrics) updateFailoverRate(ctx context.Context, service string) {
	// 尝试获取最新的故障转移率
	rate, err := p.GetFailoverRate(ctx, service)
	if err != nil {
		return
	}

	// 更新Gauge指标
	p.failoverRate.WithLabelValues(service).Set(rate)
}

// Close 关闭指标收集器
func (p *PrometheusMetrics) Close() error {
	// 最后一次推送
	if err := p.pusher.Push(); err != nil {
		return err
	}
	return nil
}

// startPushing 开始定时推送指标
func (p *PrometheusMetrics) startPushing(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := p.pusher.Push(); err != nil {
			// 记录错误但继续运行
		}
	}
}
