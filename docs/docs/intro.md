# 快速开始

## 1.基本使用

### 创建简单的rpc服务端/客户端

1. **定义服务**

```go
type User struct {
    Id   int64  `json:"id"`
    Name string `json:"name"`
}

type GetByIdReq struct {
    Id int64 `json:"id"`
}

type GetByIdResp struct {
    User *User `json:"user"`
}

// 使用结构体声明而非接口声明的方式声明方法
type UserService struct {
    GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}
```

2. **服务实现**

```go
type UserServiceImpl struct{}

func (s *UserServiceImpl) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
    // 模拟实现
    if req.Id == 123 {
        return &GetByIdResp{
            User: &User{
                Id:   req.Id,
                Name: "test",
            },
        }, nil
    }
    return &GetByIdResp{}, nil
}
```

> 注意，实现的时候不能用接口实现！由于go的反射语法不能动态实现接口类型，因此只能使用这样的方法来定义和实现服务。

3. **创建简单服务端**

* 使用`api`包创建（针对用户封装过的服务端使用）：

```go
func main() {

    // 创建服务器
    server := api.NewServer(&api.ServerOptions{
        Address: ":8000", // 服务监听地址
        SerializeType: protocol.SerializationTypeJSON, // 使用JSON序列化
    })
	
    // 注册服务
    service := &UserService{}
    err := server.Register(service)
    if err != nil {
        log.Fatalf("Failed to register service: %v", err)
    }
    
    // 启动服务器
    if err := server.Start(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
    
    // 关闭服务器
    defer func() {
        if err := server.Stop(); err != nil {
            log.Fatalf("Failed to stop server: %v", err)
        }
    }()
}
```

* 使用`rpc`包创建（底层的rpc服务端实现）：

```go
func main() {
    // 创建服务器
    server := rpc.NewServer()
    
    // 注册服务
    err := server.RegisterService(&service.UserServiceImpl{})
    if err != nil {
        log.Fatalf("Failed to register service: %v", err)
    }
    
    // 启动服务
    err = server.Start(":8080")
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
```

4. **创建简单客户端**

* 使用`api`包创建：

```go
func main() {
    // 创建客户端
    client, err := api.NewClient(&api.ClientOptions{
        Address:       ":8080",
        SerializeType: protocol.SerializationTypeJSON,
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    // 创建请求
    req := &service.GetByIdReq{Id: 123}
    resp := &service.GetByIdResp{}
    
    // 发起RPC调用
    err = client.Call(context.Background(), "User", "GetById", req, resp)
    if err != nil {
        log.Fatalf("Call failed: %v", err)
    }
    
    fmt.Printf("Got user: %+v\n", resp.User)
}
```

* 使用`rpc`包创建：

```go
func main() {

    // 创建客户端
    client, err := rpc.NewClient("localhost:8000")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    // 创建请求对象
    req := &service.GetByIdReq{Id: 123}
    
    // 执行序列化
    reqData, err := json.Marshal(req)
    if err != nil {
        log.Fatalf("Failed to serialize request: %v", err)
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // 调用远端服务
    res, err := client.CallWithContext(ctx, req, protocol.SerializationTypeJSON)
    if err != nil {
        log.Fatalf("RPC call failed: %v", err)
    }
    
    // 解析响应
    resp := &service.GetByIdResp{}
    err = json.Unmarshal(res.Data, &response)
    if err != nil {
        log.Fatalf("Failed to deserialize response: %v", err)
    }
    
    log.Printf("Response: %s", response.Message)
}
```

### 使用protobuf

1. **定义`.proto`文件**：

```protobuf
// hello.proto
syntax = "proto3";

option go_package = ".;helloworld";

service GreetService {
  rpc SayHello(HelloRequest) returns (HelloResponse) {}
}

message HelloRequest {
  string name = 1;
}

message HelloResponse {
  string message = 1;
  int64 greet_time = 2;
}
```

2. **使用protoc生成go代码**

```protobuf
protoc --go_out=. hello.proto
```

3. **实现对应的服务**

```go
type GreetServiceImpl struct{}

func (s *GreetServiceImpl) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloResponse, error) {
    return &helloworld.HelloResponse{
        Message:   "Hello, " + req.Name,
        GreetTime: time.Now().Unix(),
    }, nil
}
```

### 服务注册与发现

fyer-rpc支持etcd实现服务注册与发现功能，并提供了`Register`、`Discover`等接口，用户可以通过实现接口来使用自定义的服务注册与发现组件。

1. **服务端注册与注销**

```go
func main() {
    // 创建etcd注册中心
    registry, err := etcd.New(
        etcd.WithEndpoints([]string{"localhost:2379"}),
        etcd.WithDialTimeout(time.Second*5),
    )
    if err != nil {
        log.Fatalf("Failed to create registry: %v", err)
    }
    
    // 创建服务实例
    instance := &naming.Instance{
        ID:      "greeter-server-1",
        Service: "GreetService",
        Version: "1.0.0",
        Address: "localhost:8000",
        Status:  naming.StatusEnabled,
    }
    
    // 注册服务
    err = registry.Register(context.Background(), instance)
    if err != nil {
        log.Fatalf("Failed to register service: %v", err)
    }
    
    // 实现服务端，同上
    // ...
    
    // 在关闭服务端时注销服务
    defer registry.Deregister(context.Background(), instance)
}
```

2. **客户端服务发现**

```go
func main() {
    // 创建注册中心
    registry, err := etcd.New(
        etcd.WithEndpoints([]string{"localhost:2379"}),
        etcd.WithDialTimeout(time.Second*5),
    )
    if err != nil {
        log.Fatalf("Failed to create registry: %v", err)
    }
    
    // 创建服务发现
    disc := discovery.NewDiscovery(registry, time.Second*10)
    
    // 创建解析器
    resolver, err := discovery.NewResolver(registry, "GreetService", "1.0.0")
    if err != nil {
        log.Fatalf("Failed to create resolver: %v", err)
    }
    
    // 创建负载均衡器
    lb := discovery.NewLoadBalancer("GreetService", "1.0.0", resolver, 
        metrics.NewNoopMetrics(), balancer.Random)
    
    // 在客户端使用负载均衡器
    // 具体实现可参考example/client/client.go代码
    // ...
}
```

### 故障转移与熔断

fyer-rpc提供了具体的故障转移与熔断的简单实现，并提供了`CircuitBreaker`、`FailoverHandler`等接口，用户可以通过实现接口来使用自定义的故障转移与熔断组件。

```go
func main() {
    // 创建故障转移配置
    failoverConfig := failover.NewConfig(
        failover.WithMaxRetries(3),
        failover.WithRetryInterval(100*time.Millisecond),
        failover.WithRetryBackoff(1.5),
        failover.WithCircuitBreaker(5, 10*time.Second),
    )
    
    // 创建故障转移处理器
    failoverHandler, err := failover.NewFailoverHandler(failoverConfig)
    if err != nil {
        log.Fatalf("Failed to create failover handler: %v", err)
    }
    
    // 执行附带故障转移的操作
    instances := []*naming.Instance{
        {ID: "server1", Address: "localhost:8001"},
        {ID: "server2", Address: "localhost:8002"},
        {ID: "server3", Address: "localhost:8003"},
    }
    
    result, err := failoverHandler.Execute(context.Background(), instances, func(ctx context.Context, instance *naming.Instance) error {
        // Perform your RPC call here
        return nil
    })
    
    if err != nil {
        log.Fatalf("Operation failed: %v", err)
    }
    
    log.Printf("Operation succeeded using instance: %s", result.Instance.ID)
}
```

### 高级配置

fyer-rpc为服务端、客户端等rpc组件提供了一系列的配置选项，用户可以通过配置文件或代码来进行高级配置。下面举一些简单的例子：

1. **服务端配置**

```go
func main() {
    serverConfig := config.NewServerConfig(
        config.WithAddress(":8000"),
        config.WithNetworkConfig(100, time.Second, time.Second),
        config.WithShutdownTimeout(time.Second*10),
        config.WithServiceInfo("MyService", "1.0.0"),
        config.WithWorkerPoolSize(runtime.NumCPU()*2),
        config.WithConcurrencyLimit(1000),
    )
    
    // 使用如上配置创建服务端
}
```

2. **客户端配置**

```go
package main

import (
    "github.com/fyerfyer/fyer-rpc/config"
)

func main() {
    // 创建客户端配置
    clientConfig := config.NewClientConfig(
        config.WithPoolConfig(20, 10, time.Minute*5),
        config.WithKeepAlive(true, time.Second*30, 3),
        config.WithLoadBalancer(balancer.RoundRobin, time.Second*10),
        config.WithFailover(true, failoverConfig),
        config.WithRateLimit(100, 1000, time.Second*5),
    )
    
    // 使用如上配置创建客户端
}
```

3. **指标收集配置**

fyer-rpc集成了prometheus指标收集：

```go
func main() {
    // 创建prometheus指标收集器
    promConfig := &metrics.PrometheusConfig{
        PushGatewayURL: "localhost:9091",
        QueryURL:       "localhost:9090",
        JobName:        "fyer-rpc",
        PushInterval:   time.Second * 10,
    }
    
    metricsCollector, err := metrics.NewPrometheusMetrics(promConfig)
    if err != nil {
        log.Fatalf("Failed to create metrics collector: %v", err)
    }
    
    // 为指标收集器使用负载均衡器
    lb := discovery.NewLoadBalancer("GreetService", "1.0.0", resolver, 
        metricsCollector, balancer.FastestResponse)
    
    // 收集相应指标
    metricsCollector.RecordResponse(context.Background(), &metrics.ResponseMetric{
        ServiceName: "GreetService",
        MethodName:  "SayHello",
        Instance:    "localhost:8000",
        Duration:    time.Millisecond * 10,
        Status:      "success",
        Timestamp:   time.Now(),
    })
}
```

现在你已经对fyer-rpc框架有了基本的了解，你可以：

* 浏览其他部分的文档
* 查看示例代码，了解更复杂的使用案例
* 学习服务分组、自定义选择器等高级功能
* 将其集成到你的现有基础设施中