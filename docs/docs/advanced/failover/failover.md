# Failover

故障转移是分布式系统中重要的容错机制，当一个服务实例发生故障或不可用时，可以自动切换到其他可用实例，从而保证系统的高可用性。

## 故障转移基础

fyerrpc的故障转移系统包含如下的功能设计：

1. **快速探测**：快速准确地检测服务实例故障
2. **动态恢复**：在服务恢复后能够自动将其重新纳入可用实例池
3. **灵活策略**：提供多种故障转移策略，适应不同场景需求
4. **熔断保护**：结合熔断器模式避免持续调用不可用实例
5. **智能重试**：针对不同类型的错误采用合适的重试策略

### 故障转移流程

fyerrpc的故障转移工作流程如下：

1. **故障检测**：检测服务实例是否健康可用
2. **熔断保护**：对频繁出错的实例进行熔断
3. **实例选择**：根据策略选择可用实例
4. **调用执行**：执行远程调用操作
5. **结果处理**：根据调用结果进行状态更新和指标收集
6. **故障恢复**：定期检测并恢复不健康实例

## 核心组件

### FailoverHandler 接口

`FailoverHandler`是故障转移功能的核心接口：

```go
type FailoverHandler interface {
    // Execute 执行带故障转移的调用
    Execute(ctx context.Context, instances []*naming.Instance, operation func(context.Context, *naming.Instance) error) (*FailoverResult, error)

    // GetDetector 获取故障检测器
    GetDetector() Detector

    // GetCircuitBreaker 获取熔断器
    GetCircuitBreaker() CircuitBreaker

    // GetRetryPolicy 获取重试策略
    GetRetryPolicy() RetryPolicy

    // GetRecoveryStrategy 获取恢复策略
    GetRecoveryStrategy() RecoveryStrategy

    // GetMonitor 获取实例监控器
    GetMonitor() InstanceMonitor
}
```

### 故障转移配置

故障转移功能通过Config结构进行配置：

```go
type Config struct {
    // 重试相关配置
    MaxRetries      int           // 最大重试次数
    RetryInterval   time.Duration // 重试间隔基准时间
    RetryableErrors []string      // 可重试的错误类型列表
    RetryStrategy   string        // 重试策略: simple, exponential, jittered

    // 熔断相关配置
    CircuitBreakThreshold    int           // 熔断阈值，连续失败次数
    CircuitBreakTimeout      time.Duration // 熔断超时时间
    
    // 故障检测配置
    FailureDetectionTime time.Duration // 故障检测时间窗口
    FailureThreshold     int           // 故障阈值次数
    SuccessThreshold     int           // 成功阈值次数
    
    // 恢复策略配置
    RecoveryStrategy  string        // 恢复策略：immediate, gradual, probing
    
    // 通用配置
    FailoverStrategy string // 故障转移策略：next, random, best
}
```

可以使用函数选项模式进行配置：

```go
config := failover.NewConfig(
    failover.WithMaxRetries(3),
    failover.WithRetryInterval(100*time.Millisecond),
    failover.WithRetryBackoff(1.5, time.Second),
    failover.WithRetryJitter(0.2),
    failover.WithRetryableErrors([]string{"connection refused", "timeout"}),
    failover.WithCircuitBreaker(5, 30*time.Second),
    failover.WithFailoverStrategy("next"),
)
```

## 故障检测

故障检测负责判断服务实例是否健康，是故障转移的基础。

### Detector 接口

```go
type Detector interface {
    // Detect 检测实例是否健康
    Detect(ctx context.Context, instance *naming.Instance) (Status, error)

    // MarkFailed 标记实例为失败状态
    MarkFailed(ctx context.Context, instance *naming.Instance) error

    // MarkSuccess 标记实例为成功状态
    MarkSuccess(ctx context.Context, instance *naming.Instance) error
}
```

### 实例状态

服务实例可能处于以下状态：

```go
type Status int

const (
    StatusHealthy   Status = iota // 健康状态
    StatusUnhealthy               // 不健康状态
    StatusSuspect                 // 可疑状态，可能不健康
    StatusIsolated                // 被隔离状态
)
```

### 检测器实现

fyerrpc提供了多种故障检测器：

1. **TimeoutDetector**：基于连接超时的检测器，通过尝试TCP连接来判断实例健康状态

```go
// 创建基于超时的故障检测器
detector := failover.NewTimeoutDetector(config)

// 检测实例状态
status, err := detector.Detect(ctx, instance)
if err != nil || status != failover.StatusHealthy {
    log.Printf("Instance %s is not healthy: %v", instance.Address, err)
}
```

2. **ErrorRateDetector**：基于错误率的检测器，记录调用错误率判断实例健康状态

3. **HealthCheckDetector**：通过定期健康检查判断实例状态

## 熔断器

熔断器模式防止持续调用不健康的服务实例，提高系统的稳定性。

### CircuitBreaker 接口

```go
type CircuitBreaker interface {
    // Allow 判断请求是否允许通过
    Allow(ctx context.Context, instance *naming.Instance) (bool, error)

    // MarkSuccess 标记成功调用
    MarkSuccess(ctx context.Context, instance *naming.Instance) error

    // MarkFailure 标记失败调用
    MarkFailure(ctx context.Context, instance *naming.Instance, err error) error

    // GetState 获取熔断器状态
    GetState(instance *naming.Instance) (State, error)

    // Reset 重置熔断器状态
    Reset(instance *naming.Instance) error
}
```

### 熔断器状态

熔断器有三种状态：

```go
type State int

const (
    StateClosed   State = iota // 关闭状态，允许请求通过
    StateOpen                  // 打开状态，请求被拒绝
    StateHalfOpen              // 半开状态，允许部分请求通过以探测服务是否恢复
)
```

### 熔断器实现

fyerrpc默认使用`SimpleCircuitBreaker`，其工作流程如下：

1. **关闭状态**：所有请求正常通过，但记录失败次数
2. **打开状态**：当连续失败次数达到阈值时，熔断器打开，拒绝所有请求
3. **半开状态**：超时后熔断器进入半开状态，允许有限请求通过
4. **恢复机制**：半开状态下请求成功率达到阈值，熔断器关闭；任何失败会使熔断器重新打开

```go
// 创建熔断器
circuitBreaker := failover.NewCircuitBreaker(config)

// 判断请求是否允许通过
allow, err := circuitBreaker.Allow(ctx, instance)
if !allow {
    log.Printf("Circuit breaker is open for instance %s: %v", instance.Address, err)
    return err
}

// 根据调用结果更新熔断器
if callErr := callService(instance); callErr != nil {
    circuitBreaker.MarkFailure(ctx, instance, callErr)
} else {
    circuitBreaker.MarkSuccess(ctx, instance)
}
```

## 重试策略

重试策略定义了当调用失败时如何进行重试。

### RetryPolicy 接口

```go
type RetryPolicy interface {
    // ShouldRetry 决定是否需要重试
    ShouldRetry(ctx context.Context, attempt int, err error) bool

    // NextBackoff 计算下一次重试的等待时间
    NextBackoff(attempt int) time.Duration

    // MaxAttempts 返回最大重试次数
    MaxAttempts() int
}
```

### 重试策略实现

fyerrpc提供了多种重试策略：

1. **SimpleRetryPolicy**：简单固定间隔重试

```go
// 创建简单重试策略(最多重试3次，间隔100毫秒)
retryPolicy := failover.NewSimpleRetryPolicy(
    3,                      // 最大重试次数
    100*time.Millisecond,   // 固定重试间隔
    []string{"timeout"},    // 可重试错误类型
)
```

2. **ExponentialBackoffRetryPolicy**：指数退避重试，重试间隔随着尝试次数增加而指数增加

```go
// 创建指数退避重试策略
retryPolicy := failover.NewExponentialBackoffRetryPolicy(
    3,                     // 最大重试次数
    100*time.Millisecond,  // 初始重试间隔
    10*time.Second,        // 最大重试间隔
    2.0,                   // 乘数因子
    []string{"timeout"},   // 可重试错误类型
)
```

3. **JitteredRetryPolicy**：带随机抖动的重试策略，在指数退避基础上增加随机抖动，避免多个客户端同时重试导致的"惊群效应"

```go
// 创建带抖动的重试策略
retryPolicy := failover.NewJitteredRetryPolicy(
    3,                     // 最大重试次数
    100*time.Millisecond,  // 初始重试间隔
    10*time.Second,        // 最大重试间隔
    2.0,                   // 乘数因子
    0.2,                   // 抖动因子(±20%)
    []string{"timeout"},   // 可重试错误类型
)
```

## 恢复策略

恢复策略定义了如何将故障实例恢复到可用状态。

### RecoveryStrategy 接口

```go
type RecoveryStrategy interface {
    // CanRecover 判断实例是否可以恢复
    CanRecover(ctx context.Context, instance *naming.Instance) bool

    // Recover 恢复实例
    Recover(ctx context.Context, instance *naming.Instance) error

    // RecoveryDelay 返回恢复尝试间隔
    RecoveryDelay(instance *naming.Instance) time.Duration
}
```

### 恢复策略实现

fyerrpc提供了多种恢复策略：

1. **ImmediateRecoveryStrategy**：立即恢复策略，发现故障后立即尝试恢复

2. **GradualRecoveryStrategy**：渐进式恢复策略，在失败后使用逐渐增加的延迟尝试恢复

```go
// 创建渐进式恢复策略
recoveryStrategy := failover.NewGradualRecoveryStrategy(detector, config)
```

3. **ProbingRecoveryStrategy**：探测式恢复策略，在恢复前先发送少量请求探测服务状态

```go
// 创建探测式恢复策略
recoveryStrategy := failover.NewProbingRecoveryStrategy(detector, config)
```

## 使用故障转移

### 创建故障转移处理器

```go
// 创建配置
config := failover.NewConfig(
    failover.WithMaxRetries(3),
    failover.WithRetryInterval(100*time.Millisecond),
    failover.WithRetryBackoff(1.5, time.Second),
    failover.WithRetryJitter(0.2),
    failover.WithCircuitBreaker(5, 30*time.Second),
    failover.WithFailoverStrategy("next"),
)

// 创建故障转移处理器
handler, err := failover.NewFailoverHandler(config)
if err != nil {
    log.Fatalf("Failed to create failover handler: %v", err)
}
```

### 执行带故障转移的调用

```go
// 待调用的服务实例列表
instances := []*naming.Instance{
    {ID: "inst-1", Address: "192.168.1.101:8000", Status: naming.StatusEnabled},
    {ID: "inst-2", Address: "192.168.1.102:8000", Status: naming.StatusEnabled},
    {ID: "inst-3", Address: "192.168.1.103:8000", Status: naming.StatusEnabled},
}

// 定义操作函数
operation := func(ctx context.Context, instance *naming.Instance) error {
    // 调用服务...
    return callService(ctx, instance.Address)
}

// 执行带故障转移的调用
ctx := context.Background()
result, err := handler.Execute(ctx, instances, operation)
if err != nil {
    log.Printf("Failover call failed: %v", err)
    log.Printf("Tried %d times, failed nodes: %v", result.RetryCount, result.FailedNodes)
} else {
    log.Printf("Call succeeded using instance %s after %d retries",
        result.Instance.Address, result.RetryCount)
}
```

### 故障转移结果

故障转移操作的结果通过`FailoverResult`结构返回：

```go
type FailoverResult struct {
    Success     bool             // 是否成功
    Instance    *naming.Instance // 最终使用的实例
    RetryCount  int              // 重试次数
    Duration    time.Duration    // 操作耗时
    Error       error            // 错误信息
    FailedNodes []string         // 失败的节点列表
}
```

## 指标收集与监控

fyerrpc提供了故障转移指标收集功能，用于监控和分析系统健康状态。

### 指标收集接口

```go
type MetricsCollector interface {
    // 增加重试计数
    IncrementRetries(service string)

    // 增加故障转移计数
    IncrementFailovers(service string, fromInstance, toInstance string)

    // 记录故障检测事件
    RecordDetection(instance *naming.Instance, status Status)

    // 记录熔断器状态变更
    RecordBreaker(instance *naming.Instance, state State)

    // 记录请求延迟
    RecordLatency(instance *naming.Instance, latency time.Duration)
}
```

### 内存指标收集器

```go
// 创建内存指标收集器(保留100个延迟样本)
collector := failover.NewInMemoryMetricsCollector(100)

// 获取指标
retryCount := collector.GetRetryCount("user-service")
failoverCount := collector.GetFailoverCount("user-service")
avgLatency := collector.GetAverageLatency("instance-1")

// 获取故障检测统计
detectionStats := collector.GetDetectionStats("instance-1")
healthyCount := detectionStats[failover.StatusHealthy]
unhealthyCount := detectionStats[failover.StatusUnhealthy]

// 获取熔断器统计
breakerStats := collector.GetBreakerStats("instance-1")
openCount := breakerStats[failover.StateOpen]
```