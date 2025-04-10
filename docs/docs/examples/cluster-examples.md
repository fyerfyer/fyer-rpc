# fyerrpc Cluster Example

本文档详细介绍了 fyerrpc 框架的完整集群示例，展示了如何构建一个具有故障转移、服务发现和负载均衡功能的分布式 RPC 系统。此示例由客户端、服务器、公共组件和辅助脚本组成，是一个端到端的微服务演示。

## 概述

该示例实现了一个简单的问候服务（GreetService），包含以下组件：

1. **服务定义（helloworld）**：使用 Protocol Buffers 定义的服务接口
2. **服务器端（server）**：多个服务实例，支持故障模拟
3. **客户端（client）**：支持故障转移和负载均衡的 RPC 客户端
4. **共享组件（common）**：各组件间共享的配置和工具
5. **启动脚本（scripts）**：用于启动和停止服务集群

示例展示了以下关键功能：

- 服务发现和注册
- 客户端负载均衡
- 故障转移和重试机制
- 服务健康检查
- 熔断器和故障隔离
- 指标收集和监控

## 服务定义 (helloworld)

### 服务接口

服务接口使用 Protocol Buffers 定义：

```protobuf
// hello.proto
syntax = "proto3";

option go_package = ".;helloworld";

// GreetService 定义一个简单的问候服务
service GreetService {
  // SayHello 发送问候
  rpc SayHello(HelloRequest) returns (HelloResponse) {}
  // GetGreetStats 获取问候统计信息
  rpc GetGreetStats(StatsRequest) returns (StatsResponse) {}
}

// HelloRequest 问候请求
message HelloRequest {
  string name = 1;           // 被问候者姓名
  string greeting = 2;       // 自定义问候语(可选)
}

// HelloResponse 问候响应
message HelloResponse {
  string message = 1;        // 问候消息
  int64 greet_time = 2;      // 问候时间戳
}

// StatsRequest 统计请求
message StatsRequest {
  string name = 1;          // 查询指定用户的统计(可选)
}

// StatsResponse 统计响应
message StatsResponse {
  int64 total_greets = 1;   // 总问候次数
  map<string, int64> greets_by_name = 2;  // 各用户的问候次数
  int64 last_greet_time = 3;    // 最后一次问候时间
}
```

### 服务实现

服务的基本实现逻辑：

```go
// hello.go
package helloworld

import (
	"context"
	"sync"
	"time"
)

// GreetServiceImpl 问候服务实现
type GreetServiceImpl struct {
	mu sync.RWMutex
	// 统计信息
	totalGreets   int64
	greetsByName  map[string]int64
	lastGreetTime int64
}

// SayHello 实现问候方法
func (s *GreetServiceImpl) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	// 更新统计信息并返回问候消息
	// ...
}

// GetGreetStats 实现统计方法
func (s *GreetServiceImpl) GetGreetStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	// 返回统计信息
	// ...
}
```

## 公共组件 (common)

### 客户端和服务器配置

config.go 定义了客户端和服务器的配置结构：

```go
// ServerConfig 服务器配置
type ServerConfig struct {
	// 基本配置
	ID      string // 服务器ID
	Address string // 服务地址
	Port    int    // 服务端口

	// 故障模拟相关
	FailAfter    int           // 在处理这么多个请求后故障
	FailDuration time.Duration // 故障持续时间
	FailRate     float64       // 随机故障概率 (0-1)
}

// ClientConfig 客户端配置
type ClientConfig struct {
	// 基本配置
	ServerAddresses []string      // 服务器地址列表
	Timeout         time.Duration // 请求超时时间

	// 故障转移配置
	FailoverConfig *failover.Config // 故障转移配置
	EnableFailover bool             // 是否启用故障转移
}
```

### 简单指标收集器

metrics.go 实现了一个简单的内存指标收集器：

```go
// SimpleMetrics 是一个简单的指标收集实现，用于示例
type SimpleMetrics struct {
	// 计数器
	requestCount  int64 // 请求总数
	successCount  int64 // 成功请求数
	failureCount  int64 // 失败请求数
	retryCount    int64 // 重试次数
	failoverCount int64 // 故障转移次数
	circuitBreaks int64 // 熔断次数

	// 实例级别统计
	instanceStats map[string]*InstanceStat

	// 响应时间和事件记录
	// ...
}

// 各种记录和获取指标的方法
// ...
```

## 服务器端 (server)

服务器端实现了支持故障模拟的 RPC 服务，主要包括以下组件：

### 主程序 (main.go)

主程序负责解析命令行参数、创建服务器实例、注册服务并启动监听：

```go
func main() {
	// 解析命令行参数
	flag.Parse()

	// 创建 etcd 注册中心
	registry, err := etcd.New(/* ... */)
	
	// 根据命令行参数创建服务器配置
	serverConfig := &common.ServerConfig{
		ID:           *id,
		Address:      "localhost",
		Port:         *port,
		FailAfter:    *failAfter,
		FailDuration: *failDuration,
		FailRate:     *failRate,
	}

	// 创建 Prometheus 指标服务
	// ...

	// 创建 RPC 服务器
	server := rpc.NewServer()

	// 创建并启动 Greet 服务
	greetServer := NewGreetServer(registry, serverConfig)
	
	// 启动服务注册和健康检查
	// ...

	// 等待终止信号并优雅关闭
	// ...
}
```

### 健康检测器 (detector.go)

健康检测器负责模拟服务故障并提供健康检查接口：

```go
// HealthDetector 健康检测器，管理服务健康状态并模拟故障情况
type HealthDetector struct {
	config        *common.ServerConfig
	requestCount  int64  // 请求计数器
	failureTime   *int64 // 故障开始时间（如果当前处于故障状态）
	statusHandler http.Handler
	metrics       *common.SimpleMetrics
	mu            sync.RWMutex
}

// IsHealthy 检查服务是否健康
func (d *HealthDetector) IsHealthy() bool {
	// 根据配置模拟故障
	// 1. 检查随机故障率
	// 2. 检查请求数是否达到故障阈值
	// 3. 检查故障持续时间是否已过
	// ...
}

// 提供健康检查和指标 HTTP 接口
// ...
```

### 服务实现 (service.go)

封装了 GreetService 的服务实现，添加了故障模拟和监控功能：

```go
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

// SayHello 包装原始的 SayHello 方法，添加故障模拟逻辑
func (s *GreetServer) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloResponse, error) {
	// 增加请求计数
	reqCount := atomic.AddInt64(&s.requestCount, 1)

	// 检查是否应该模拟故障
	isHealthy := s.detector.IsHealthy()
	if !isHealthy {
		return nil, fmt.Errorf("service %s is currently unavailable", s.config.ID)
	}

	// 记录请求开始时间和指标
	// 调用原始服务实现
	// ...
}
```

## 客户端 (client)

客户端实现了支持故障转移的 RPC 调用功能，主要包括以下组件：

### 故障转移管理器 (failover.go)

故障转移管理器负责处理服务实例故障并自动切换到健康实例：

```go
// FailoverManager 管理客户端的故障转移功能
type FailoverManager struct {
	handler        *failover.DefaultFailoverHandler // 故障转移处理器
	config         *failover.Config                 // 故障转移配置
	metrics        *common.SimpleMetrics            // 指标收集器
	serverList     []*naming.Instance               // 服务器实例列表
	activeInstance *naming.Instance                 // 当前活跃的实例
	mu             sync.RWMutex
}

// ExecuteRPC 执行带故障转移的RPC调用
func (fm *FailoverManager) ExecuteRPC(ctx context.Context, serviceName, methodName string, req interface{}, resp interface{}) error {
	// 复制实例列表
	// 定义RPC调用操作
	// 执行带故障转移的调用
	// 记录故障和恢复事件
	// ...
}
```

### 客户端实现 (client.go)

封装了 GreetService 的客户端调用逻辑：

```go
// GreetClient 包装了问候服务客户端的实现
type GreetClient struct {
	balancer        *discovery.LoadBalancer
	metrics         metrics.Metrics
	discovery       discovery.Discovery
	failoverManager *FailoverManager     // 故障转移管理器
	config          *common.ClientConfig // 客户端配置
}

// SayHello 调用问候服务
func (c *GreetClient) SayHello(ctx context.Context, name string, greeting string) (*helloworld.HelloResponse, error) {
	// 构造请求
	// 选择调用方式（故障转移或负载均衡）
	// 记录调用结果
	// ...
}
```

### 主程序 (main.go)

主程序演示了各种故障转移场景：

```go
func main() {
	// 创建带故障转移功能的配置
	serverAddresses := []string{
		"localhost:8001",
		"localhost:8002",
		"localhost:8003",
	}
	clientConfig := CreateDefaultClientConfig(serverAddresses)

	// 创建带故障转移功能的客户端
	client, err := NewGreetClient("GreetService", clientConfig)
	
	// 启动健康监测器
	// ...

	// 测试一般调用
	testBasicCalls(ctx, client)

	// 测试故障转移
	testFailover(ctx, client)

	// 测试熔断器
	testCircuitBreaker(ctx, client)

	// 测试并发调用
	testConcurrentCalls(ctx, client)

	// 显示最终指标
	showFailoverMetrics(client)
}
```

## 启动脚本 (scripts)

提供了启动和停止服务集群的脚本：

### start_cluster.sh (Linux/macOS)

Linux/macOS 环境下启动多个服务实例的脚本：

```bash
#!/bin/bash
# 启动多个服务器实例，用于演示故障转移功能

echo "Starting server cluster for failover demonstration..."

# 设置基础端口号
BASE_PORT=8001
SERVERS_COUNT=3

# Server A - 处理100个请求后故障
echo "Starting Server A (Port $BASE_PORT) - Fails after 100 requests for 10s"
cd ../server && go run . -port=$BASE_PORT -id=server-A -fail-after=100 -fail-duration=10s > ../scripts/logs/server_a.log 2>&1 &

# Server B - 10%概率随机故障
PORT_B=$((BASE_PORT+1))
echo "Starting Server B (Port $PORT_B) - 10% random failure rate"
cd ../server && go run . -port=$PORT_B -id=server-B -fail-rate=0.1 > ../scripts/logs/server_b.log 2>&1 &

# Server C - 正常运行
PORT_C=$((BASE_PORT+2))
echo "Starting Server C (Port $PORT_C) - Normal operation"
cd ../server && go run . -port=$PORT_C -id=server-C > ../scripts/logs/server_c.log 2>&1 &

# 启动客户端示例
echo "Starting failover client demo..."
cd ../client && go run . > ../scripts/logs/client.log 2>&1 &
```

### start_cluster.bat (Windows)

Windows 环境下启动多个服务实例的脚本：

```batch
@echo off
REM 启动多个服务器实例，用于演示故障转移功能
echo Starting server cluster for failover demonstration...

REM 设置基础端口号
set BASE_PORT=8001
set SERVERS_COUNT=3

REM Server A - 处理100个请求后故障
start "Server A (Port %BASE_PORT%)" cmd /c "cd ..\server && go run . -port=%BASE_PORT% -id=server-A -fail-after=100 -fail-duration=10s"

REM Server B - 10%概率随机故障
set /a "PORT_B=%BASE_PORT%+1"
start "Server B (Port %PORT_B%)" cmd /c "cd ..\server && go run . -port=%PORT_B% -id=server-B -fail-rate=0.1"

REM Server C - 正常运行
set /a "PORT_C=%BASE_PORT%+2"
start "Server C (Port %PORT_C%)" cmd /c "cd ..\server && go run . -port=%PORT_C% -id=server-C"

REM 启动客户端示例
start "Failover Client Demo" cmd /c "cd ..\client && go run ."
```

## 功能演示

整个示例演示了以下核心功能：

### 1. 基本 RPC 调用

客户端向服务器发送基本的 RPC 请求，展示正常调用流程：

```go
// 测试基本调用
func testBasicCalls(ctx context.Context, client *GreetClient) {
    for i := 0; i < 3; i++ {
        resp, err := client.SayHello(ctx, fmt.Sprintf("User%d", i), "Hello")
        if err != nil {
            log.Printf("Error: %v", err)
        } else {
            log.Printf("Response: %s", resp.Message)
        }
    }
}
```

### 2. 故障转移

通过快速发送多个请求触发 Server A 的故障模拟，然后观察系统如何自动切换到其他可用服务器：

```go
// 测试故障转移
func testFailover(ctx context.Context, client *GreetClient) {
    // 快速发送多个请求，触发服务器故障
    for i := 0; i < 120; i++ {
        resp, err := client.SayHello(ctx, fmt.Sprintf("User%d", i), "Hello")
        // ...
    }

    // 等待一下，让服务器恢复
    fmt.Println("\n* Waiting for Server A to recover...")
    time.Sleep(12 * time.Second)

    // 验证恢复后是否正常
    // ...
}
```

### 3. 熔断器

演示熔断器功能，当服务器持续失败时，熔断器开启以避免持续调用不可用服务：

```go
// 测试熔断器
func testCircuitBreaker(ctx context.Context, client *GreetClient) {
    // 创建一个短超时的上下文，强制产生超时错误以触发熔断
    timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
    defer cancel()

    // 尝试多次请求，触发熔断
    for i := 0; i < 5; i++ {
        _, err := client.SayHello(timeoutCtx, "CircuitBreakerTest", "Hello")
        // ...
    }

    // 等待熔断恢复
    // ...
}
```

### 4. 并发调用

测试系统在并发请求下的行为：

```go
// 测试并发调用
func testConcurrentCalls(ctx context.Context, client *GreetClient) {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            resp, err := client.SayHello(ctx, fmt.Sprintf("ConcurrentUser%d", id), "Hello")
            // ...
        }(i)
    }
    wg.Wait()
}
```

### 5. 指标收集

整个过程中收集各种性能指标和事件数据，最后展示汇总结果：

```go
// 显示故障转移指标
func showFailoverMetrics(client *GreetClient) {
    metrics := client.GetFailoverMetrics()
    
    total, success, failure := metrics.GetRequestCount()
    retryCount := metrics.GetRetryCount()
    failoverCount := metrics.GetFailoverCount()
    circuitBreaks := metrics.GetCircuitBreaks()
    avgResponseTime := metrics.GetAvgResponseTime()

    fmt.Println("\n=== Failover Metrics ===")
    fmt.Printf("Total requests: %d (Success: %d, Failure: %d)\n", total, success, failure)
    fmt.Printf("Retry count: %d\n", retryCount)
    fmt.Printf("Failover count: %d\n", failoverCount)
    // ...
}
```

## 运行示例

要运行此完整示例，请按以下步骤操作：

1. 确保已安装并启动 etcd：

```bash
# 使用 Docker 运行 etcd
docker run -d --name etcd \
  -p 2379:2379 -p 2380:2380 \
  quay.io/coreos/etcd:v3.4.15 \
  /usr/local/bin/etcd \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-client-urls http://0.0.0.0:2379
```

2. 启动服务集群：

```bash
# Linux/macOS
cd scripts
./start_cluster.sh

# Windows
cd scripts
start_cluster.bat
```

3. 观察结果：

运行后，客户端会自动执行一系列测试，包括基本调用、故障转移测试、熔断器测试和并发调用测试，然后显示收集的指标数据。

可以通过查看各个服务器的健康检查端点来监控服务状态：
- Server A: http://localhost:18001/health
- Server B: http://localhost:18002/health
- Server C: http://localhost:18003/health
