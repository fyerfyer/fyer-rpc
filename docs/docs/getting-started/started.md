# Getting Started

下面通过一个简单的示例了解如何使用fyerrpc创建、启动和调用一个RPC服务。

## 创建服务接口

fyerrpc的服务接口创建有如下规范：

* 服务实现的命名必须按照`ServiceName+"Impl"`的格式，比如对于`GreeterService`服务，服务实现类命名为`GreeterServiceImpl`。

> 这是因为fyerrpc的客户端在获取服务类型时使用了如下代码：

```go
// 使用结构体的基础名称（去掉Impl后缀）作为服务名
serviceName := serviceType.Name()
if len(serviceName) > 4 && serviceName[len(serviceName)-4:] == "Impl" {
    serviceName = serviceName[:len(serviceName)-4]
}
```

- RPC方法实现的第一个参数必须是`context.Context`，第二个参数必须是请求对象的指针
- 返回值必须是响应对象的指针和错误

---

```go
// 请求结构体
type HelloRequest struct {
    Name string `json:"name"`
}

// 响应结构体
type HelloResponse struct {
    Message string `json:"message"`
}

// 定义服务结构体
type GreeterServiceImpl struct{}

// SayHello 方法实现
func (s *GreeterServiceImpl) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
    return &HelloResponse{
        Message: fmt.Sprintf("Hello, %s!", req.Name),
    }, nil
}

// SayGoodbye 方法实现
func (s *GreeterServiceImpl) SayGoodbye(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
    return &HelloResponse{
        Message: fmt.Sprintf("Goodbye, %s!", req.Name),
    }, nil
}
```

## 创建服务端

```go
func main() {
    // 配置日志
    utils.SetDefaultLogger(utils.NewLogger(utils.InfoLevel, os.Stdout))

    // 创建服务器配置
    options := &api.ServerOptions{
        Address:       ":8000",                        // 服务监听地址
        SerializeType: protocol.SerializationTypeJSON, // 使用JSON序列化
    }

    // 创建服务器
    server := api.NewServer(options)

    // 注册服务
    greeter := &GreeterService{}
    err := server.Register(greeter)
    if err != nil {
        log.Fatalf("Failed to register service: %v", err)
    }

    // 启动服务器
    log.Println("Starting RPC server on", options.Address)
    if err := server.Start(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }

    // 等待终止信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh

    log.Println("Shutting down server...")
    server.Stop()
}
```

## 创建客户端

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/fyerfyer/fyer-rpc/api"
    "github.com/fyerfyer/fyer-rpc/protocol"
)

func main() {
    // 创建客户端配置
    options := &api.ClientOptions{
        Address:       "localhost:8000",               // 服务器地址
        Timeout:       time.Second * 5,                // 请求超时
        SerializeType: protocol.SerializationTypeJSON, // 使用JSON序列化
    }

    // 创建客户端
    client, err := api.NewClient(options)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

    // 创建请求对象
    request := &HelloRequest{
        Name: "World",
    }

    // 创建响应对象
    response := &HelloResponse{}

    // 调用远程服务
    // 注意：服务名称就是结构体的名称 "GreeterService"
    err = client.Call(context.Background(), "GreeterService", "SayHello", request, response)
    if err != nil {
        log.Fatalf("RPC call failed: %v", err)
    }

    // 打印响应
    log.Printf("SayHello response: %s", response.Message)

    // 调用另一个方法
    err = client.Call(context.Background(), "GreeterService", "SayGoodbye", request, response)
    if err != nil {
        log.Fatalf("RPC call failed: %v", err)
    }

    // 打印响应
    log.Printf("SayGoodbye response: %s", response.Message)
}
```

## 运行示例

1. 在一个终端中启动服务端：

```bash
go run server.go
```

2. 在另一个终端中运行客户端：

```bash
go run client.go
```

## 下一步

这个简单的示例展示了fyerrpc的基本使用方法。在实际应用中，您可能需要更多高级特性：

- 服务注册与发现：与etcd等注册中心集成，实现服务的自动发现
- 负载均衡：实现请求在多个服务实例间的分配
- 故障转移：当服务实例故障时，自动重试其他实例
- 熔断保护：防止雪崩效应
- 监控指标：收集调用统计信息以帮助分析性能问题

请参考详细文档了解这些高级特性的使用方法。

## 完整示例

完整示例代码可以在quickstart目录中找到。这个示例提供了一个功能完整的RPC服务，包括服务定义、服务端实现和客户端调用。

> 由于`server.go`、`client.go`和`service.go`定义在同一个目录下，运行的时候需要执行不太一样的指令：

```bash
go run service.go server.go
# client类似
```