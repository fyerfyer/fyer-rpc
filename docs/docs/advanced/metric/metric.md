# Monitor & Metrics

监控是微服务架构中的关键组件，帮助开发者了解系统运行状况、性能瓶颈以及问题排查。

## 指标收集

### 关键指标

fyerrpc监控系统收集并暴露以下关键指标：

1. **响应时间**：RPC调用的延迟时间分布
2. **请求计数**：成功和失败的请求总数
3. **错误率**：RPC调用的错误比率
4. **故障转移**：故障转移事件计数和相关统计
5. **熔断状态**：熔断器状态变更和统计
6. **重试次数**：RPC请求的重试统计

### 指标接口

fyerrpc使用`metrics.Metrics`接口定义了指标收集的标准行为：

```go
type Metrics interface {
    // RecordResponse 记录响应时间
    RecordResponse(ctx context.Context, metric *ResponseMetric) error

    // GetLatency 获取指定服务实例的平均响应时间
    GetLatency(ctx context.Context, service, instance string) (time.Duration, error)

    // GetServiceLatency 获取服务所有实例的平均响应时间
    GetServiceLatency(ctx context.Context, service string) (map[string]time.Duration, error)

    // RecordFailover 记录故障转移事件
    RecordFailover(ctx context.Context, service, fromInstance, toInstance string) error

    // RecordCircuitBreak 记录熔断事件
    RecordCircuitBreak(ctx context.Context, service, instance string, state string) error

    // RecordRetry 记录重试事件
    RecordRetry(ctx context.Context, service, instance string, attempt int) error

    // GetFailoverRate 获取故障转移率
    GetFailoverRate(ctx context.Context, service string) (float64, error)

    // Close 关闭指标收集器
    Close() error
}
```

### 响应指标

每个RPC调用的指标数据使用`ResponseMetric`结构体表示：

```go
type ResponseMetric struct {
    ServiceName string            // 服务名称
    MethodName  string            // 方法名称
    Instance    string            // 实例地址
    Duration    time.Duration     // 响应时间
    Status      string            // 调用状态 (success/error)
    Timestamp   time.Time         // 记录时间戳
    Tags        map[string]string // 额外的标签信息
}
```

### 内存指标收集器

当不需要外部监控系统时，fyerrpc提供了内置的内存指标收集器，适用于测试、调试和小型部署场景：

```go
// 创建内存指标收集器
collector := failover.NewInMemoryMetricsCollector(100) // 保留100个延迟样本

// 使用内存指标收集器
result, err := handler.Execute(ctx, instances, operation)
collector.RecordLatency(result.Instance, result.Duration)

// 获取统计信息
avgLatency := collector.GetAverageLatency("instance-1")
failoverCount := collector.GetFailoverCount("user-service")
```

### 空操作指标收集器

fyerrpc还提供了空操作指标收集器(`NoopMetrics`)，用于禁用指标收集：

```go
// 创建空操作指标收集器
metrics := metrics.NewNoopMetrics()

// 配置客户端使用空操作指标收集器
config := &balancer.Config{
    Type:           balancer.Random,
    MetricsClient:  metrics,
    UpdateInterval: 10,
    RetryTimes:     3,
}
```

## Prometheus集成

### Prometheus简介

[Prometheus](https://prometheus.io/)是一个开源的监控和告警系统，专为微服务架构设计，具有以下特点：

- 多维度数据模型和查询语言
- 无需依赖外部存储的时序数据库
- 支持通过HTTP协议进行拉取式数据收集
- 支持通过中间网关进行推送式数据收集
- 多种图形和仪表盘支持

### PrometheusMetrics

fyerrpc提供了`PrometheusMetrics`实现，将框架指标直接集成到Prometheus生态：

```go
// 创建Prometheus指标收集器
metricsClient, err := metrics.NewPrometheusMetrics(&metrics.PrometheusConfig{
    PushGatewayURL: "http://localhost:9091", // Prometheus推送网关地址
    QueryURL:       "http://localhost:9090", // Prometheus查询地址
    JobName:        "fyerrpc",               // 作业名称
    PushInterval:   time.Second * 10,        // 推送间隔
})
if err != nil {
    log.Fatalf("Failed to create metrics client: %v", err)
}
defer metricsClient.Close()
```

### 关键指标定义

fyerrpc为Prometheus定义了以下关键指标：

```go
// 响应时间直方图
responseTime := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "rpc_response_time_seconds",
        Help:    "RPC response time in seconds",
        Buckets: prometheus.DefBuckets,
    },
    []string{"service", "method", "instance", "status"},
)

// 请求总数计数器
requestTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "rpc_requests_total",
        Help: "Total number of RPC requests",
    },
    []string{"service", "method", "instance", "status"},
)

// 故障转移计数器
failoverTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "rpc_failover_total",
        Help: "Total number of failover events",
    },
    []string{"service", "from_instance", "to_instance"},
)

// 熔断器事件计数器
circuitBreaks := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "rpc_circuit_breaks_total",
        Help: "Total number of circuit breaker state changes",
    },
    []string{"service", "instance", "state"},
)

// 重试计数器
retryTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "rpc_retry_total",
        Help: "Total number of retry attempts",
    },
    []string{"service", "instance", "attempt"},
)

// 故障转移率
failoverRate := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "rpc_failover_rate",
        Help: "Rate of failovers per service",
    },
    []string{"service"},
)
```

### 使用PrometheusMetrics

将Prometheus指标收集器与fyerrpc组件集成：

```go
// 与负载均衡器集成
lb, err := balancer.Build(&balancer.Config{
    Type:           balancer.FastestResponse,
    MetricsClient:  metricsClient,
    UpdateInterval: 30,
    RetryTimes:     3,
})

// 与故障转移处理器集成
handler, err := failover.NewFailoverHandler(&failover.Config{
    MaxRetries:      3,
    RetryInterval:   time.Millisecond * 100,
    EnableMetrics:   true,
    FailoverStrategy: "next",
})

// 记录RPC调用指标
metricsClient.RecordResponse(ctx, &metrics.ResponseMetric{
    ServiceName: "user-service",
    MethodName:  "GetUser",
    Instance:    "10.0.0.1:8080",
    Duration:    time.Millisecond * 50,
    Status:      "success",
    Timestamp:   time.Now(),
})
```

### 查询指标数据

使用PrometheusMetrics查询性能指标：

```go
// 获取指定服务实例的平均响应时间
latency, err := metricsClient.GetLatency(ctx, "user-service", "10.0.0.1:8080")
if err != nil {
    log.Printf("Failed to get latency: %v", err)
} else {
    log.Printf("Average latency: %v", latency)
}

// 获取服务所有实例的平均响应时间
latencyMap, err := metricsClient.GetServiceLatency(ctx, "user-service")
if err != nil {
    log.Printf("Failed to get service latencies: %v", err)
} else {
    for instance, latency := range latencyMap {
        log.Printf("Instance %s average latency: %v", instance, latency)
    }
}

// 获取故障转移率
rate, err := metricsClient.GetFailoverRate(ctx, "user-service")
if err != nil {
    log.Printf("Failed to get failover rate: %v", err)
} else {
    log.Printf("Failover rate: %.2f%%", rate*100)
}
```

## 监控配置

### 客户端监控配置

在客户端配置中启用监控：

```go
// 创建客户端配置
config := config.NewClientConfig(
    // 其他配置...
    
    // 配置指标收集和监控
    config.WithMetrics(true, metricsClient),
    config.WithMonitoringInterval(time.Second * 5),
)

// 创建客户端
client, err := api.NewClient(&api.ClientOptions{
    Address:       "localhost:8000",
    Metrics:       metricsClient,
    EnableMetrics: true,
})
```

### 服务端监控配置

在服务端配置中启用监控：

```go
// 创建服务器配置
options := &api.ServerOptions{
    Address:        ":8000",
    EnableRegistry: true,
    Registry:       registry,
    ServiceName:    "greeter-service",
    ServiceVersion: "1.0.0",
    
    // 监控配置
    Metrics:        metricsClient,
    EnableMetrics:  true,
    MetricsPath:    "/metrics",  // 暴露Prometheus指标的HTTP路径
}

// 创建服务器
server := api.NewServer(options)
```

## 集成Prometheus监控栈

fyerrpc可以与完整的Prometheus监控栈集成，包括Grafana、AlertManager等：

### Prometheus配置

在Prometheus配置文件中添加fyerrpc服务的抓取配置：

```yaml
scrape_configs:
  - job_name: 'fyerrpc'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:8080']
```

如果使用推送网关，则配置如下：

```yaml
scrape_configs:
  - job_name: 'pushgateway'
    scrape_interval: 5s
    static_configs:
      - targets: ['localhost:9091']
    honor_labels: true
```

## 自定义指标收集器

您可以通过实现`metrics.Metrics`接口创建自定义的指标收集器：

```go
// 自定义指标收集器实现
type MyCustomMetrics struct {
    // 您的指标收集器字段
}

// RecordResponse 记录响应时间
func (m *MyCustomMetrics) RecordResponse(ctx context.Context, metric *metrics.ResponseMetric) error {
    // 实现响应时间记录逻辑
    return nil
}

// GetLatency 获取指定服务实例的平均响应时间
func (m *MyCustomMetrics) GetLatency(ctx context.Context, service, instance string) (time.Duration, error) {
    // 实现获取延迟时间的逻辑
    return time.Millisecond * 100, nil
}

// 实现其他接口方法...

// 使用自定义指标收集器
myMetrics := &MyCustomMetrics{}
client, err := api.NewClient(&api.ClientOptions{
    Address: "localhost:8000",
    Metrics: myMetrics,
})
```

## 示例：完整监控配置

下面是一个集成Prometheus的完整监控配置示例：

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/fyerfyer/fyer-rpc/api"
	"github.com/fyerfyer/fyer-rpc/discovery/metrics"
)

func main() {
	// 创建Prometheus指标收集器
	metricsClient, err := metrics.NewPrometheusMetrics(&metrics.PrometheusConfig{
		PushGatewayURL: "http://prometheus-pushgateway:9091",
		QueryURL:       "http://prometheus:9090",
		JobName:        "fyerrpc-user-service",
		PushInterval:   time.Second * 5,
	})
	if err != nil {
		log.Fatalf("Failed to create metrics client: %v", err)
	}
	defer metricsClient.Close()

	// 创建服务器配置
	options := &api.ServerOptions{
		Address:       ":8080",
		ServiceName:   "user-service",
		ServiceVersion: "1.0.0",
		
		// 监控配置
		Metrics:       metricsClient,
		EnableMetrics: true,
		MetricsPath:   "/metrics",
	}

	// 创建服务器
	server := api.NewServer(options)

	// 注册服务
	server.Register(&UserService{})

	// 启动服务器
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 等待信号停止服务器
	// ...
}
```
