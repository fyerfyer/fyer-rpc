# Client

客户端是fyerrpc框架的核心组件之一，负责向服务端发起RPC请求并处理响应。

## 基础客户端

### Client接口

fyerrpc框架通过`api.Client`接口定义了客户端的基本行为：

```go
type Client interface {
    // Call 同步调用远程服务
    Call(ctx context.Context, service, method string, req interface{}, resp interface{}) error

    // Close 关闭客户端连接
    Close() error
}
```

### 创建客户端实例

```go
import "github.com/fyerfyer/fyer-rpc/api"

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
```

如果不提供配置项，客户端将使用默认配置：

```go
// 使用默认配置创建客户端
client, err := api.NewClient(&api.ClientOptions{
    Address: "localhost:8000",  // 地址是必须提供的
})
```

### 调用远程服务

使用客户端调用远程服务：

```go
// 定义请求和响应结构体
type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

// 创建请求对象
request := &HelloRequest{Name: "World"}

// 创建响应对象
response := &HelloResponse{}

// 调用远程服务
ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
defer cancel()

err = client.Call(ctx, "GreeterService", "SayHello", request, response)
if err != nil {
    log.Fatalf("RPC call failed: %v", err)
}

fmt.Printf("Response: %s\n", response.Message)
```

## 底层客户端实现

除了高级的`api.Client`接口，fyerrpc还提供了更底层的`rpc.Client`实现，为需要更多控制和定制的场景提供支持。

### 创建底层客户端

```go
import "github.com/fyerfyer/fyer-rpc/rpc"

// 创建客户端
client, err := rpc.NewClient("localhost:8000")
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
defer client.Close()
```

底层客户端支持通过选项模式配置高级功能：

```go
// 创建带故障转移功能的客户端
config := &failover.Config{
    MaxAttempts:      3,
    RetryInterval:    time.Millisecond * 100,
    RetryableErrors:  []string{"connection refused", "timeout"},
    FailureThreshold: 5,
    SuccessThreshold: 2,
    ResetInterval:    time.Minute,
}

client, err := rpc.NewClient("localhost:8000", 
    rpc.WithFailover(config),
)
```

### 底层调用方法

`rpc.Client`提供了以下调用方法：

```go
// 基本RPC调用
func (c *Client) Call(serviceName, methodName string, args interface{}) ([]byte, error)

// 带超时的RPC调用
func (c *Client) CallWithTimeout(ctx context.Context, serviceName, methodName string, args interface{}) ([]byte, error)

// 带故障转移的RPC调用
func (c *Client) CallWithFailover(ctx context.Context, serviceName, methodName string, args interface{}, instances []*naming.Instance) ([]byte, error)
```

底层调用示例：

```go
// 创建请求对象
request := &HelloRequest{Name: "World"}

// 序列化请求参数（底层客户端不会自动序列化）
serializer := codec.GetCodec(codec.JSON)
reqBytes, err := serializer.Encode(request)
if err != nil {
    log.Fatalf("Failed to encode request: %v", err)
}

// 带超时的RPC调用
ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
defer cancel()

respBytes, err := client.CallWithTimeout(ctx, "GreeterService", "SayHello", request)
if err != nil {
    log.Fatalf("RPC call failed: %v", err)
}

// 手动解析响应
response := &HelloResponse{}
err = serializer.Decode(respBytes, response)
if err != nil {
    log.Fatalf("Failed to decode response: %v", err)
}
```

## 客户端配置

fyerrpc提供了丰富的客户端配置选项，通过`config.ClientConfig`进行设置。

### ClientConfig

`config.ClientConfig`包含了详细的客户端配置项：

```go
type ClientConfig struct {
    *CommonConfig // 继承通用配置

    // 连接相关配置
    PoolSize        int           // 连接池大小
    MaxIdle         int           // 最大空闲连接数
    IdleTimeout     time.Duration // 空闲连接超时时间
    KeepAlive       bool          // 是否保持连接活跃
    KeepAliveTime   time.Duration // 连接保活时间
    KeepAliveCount  int           // 保活探测次数
    KeepAliveIdle   time.Duration // 连接空闲多久开始保活探测
    ConnectionLimit int           // 单个地址最大连接数

    // 负载均衡相关配置
    LoadBalanceType    balancer.BalancerType // 负载均衡类型
    UpdateInterval     time.Duration         // 服务发现更新间隔
    EnableConsistentLB bool                  // 是否启用一致性负载均衡

    // 故障转移配置
    EnableFailover  bool             // 是否启用故障转移
    FailoverConfig  *failover.Config // 故障转移配置
    FailoverTimeout time.Duration    // 故障转移超时时间

    // 限流相关配置
    MaxConcurrentRequests int           // 最大并发请求数
    MaxQPS                int           // 每秒最大请求数
    RequestTimeout        time.Duration // 请求超时时间
}
```

### 默认配置

fyerrpc为ClientConfig提供了合理的默认值：

```go
var DefaultClientConfig = &ClientConfig{
    CommonConfig: DefaultCommonConfig,

    // 连接相关默认配置
    PoolSize:        10,
    MaxIdle:         5,
    IdleTimeout:     time.Minute * 5,
    KeepAlive:       true,
    KeepAliveTime:   time.Second * 30,
    KeepAliveCount:  3,
    KeepAliveIdle:   time.Second * 60,
    ConnectionLimit: 100,

    // 负载均衡相关默认配置
    LoadBalanceType:    balancer.Random,
    UpdateInterval:     time.Second * 10,
    EnableConsistentLB: false,

    // 故障转移默认配置
    EnableFailover:  true,
    FailoverConfig:  failover.DefaultConfig,
    FailoverTimeout: time.Second * 10,

    // 限流相关默认配置
    MaxConcurrentRequests: 100,
    MaxQPS:                1000,
    RequestTimeout:        time.Second * 5,
}
```

### 配置选项函数

fyerrpc采用选项模式，通过函数链式调用进行配置：

```go
// 创建客户端配置
config := config.NewClientConfig(
    // 设置连接池配置
    config.WithPoolConfig(20, 10, time.Minute*10),
    
    // 设置连接保活配置
    config.WithKeepAlive(true, time.Second*60, time.Minute*2, 5),
    
    // 设置负载均衡配置
    config.WithLoadBalancer(balancer.RoundRobin, time.Second*15, false),
    
    // 设置故障转移配置
    config.WithFailover(true, failover.DefaultConfig, time.Second*15),
    
    // 设置限流配置
    config.WithRateLimit(200, 2000, time.Second*10),
)

// 应用通用配置
config.Apply(
    config.WithSerialization(config.SerializationJSON),
    config.WithTimeouts(time.Second*3, time.Second*10),
    config.WithRetry(3, time.Millisecond*200, []string{"connection_refused", "timeout"}),
)

// 初始化配置
config.Init()
```

## 连接池

fyerrpc客户端支持连接池管理，有效提高了资源利用率和性能。

### ConnPool

`rpc.ConnPool`类实现了连接池功能：

```go
type ConnPool struct {
    mu          sync.Mutex
    address     string
    maxIdle     int
    idleTimeout time.Duration
    conns       chan *Client
}
```

### 创建连接池

```go
// 创建连接池
pool := rpc.NewConnPool(
    "localhost:8000",  // 服务器地址
    10,               // 最大空闲连接数
    time.Minute*5,    // 空闲连接超时时间
)
defer pool.Close()
```

### 使用连接池

```go
// 从连接池获取连接
client, err := pool.Get()
if err != nil {
    log.Fatalf("Failed to get client from pool: %v", err)
}

// 使用客户端...
respBytes, err := client.Call("GreeterService", "SayHello", request)

// 归还连接到池中
pool.Put(client)
```

连接池工作流程：

1. 当调用`Get()`时，如果池中有可用连接，返回一个已有的连接
2. 如果池为空，会创建一个新的连接
3. 当调用`Put(client)`时，如果池未满，连接会被放回池中供后续使用
4. 如果池已满，连接会被直接关闭
5. 当池关闭时，所有池中连接会被关闭释放

## 动态代理

fyerrpc提供了动态代理功能，让开发者可以像调用本地方法一样调用远程服务，不需要手动处理序列化和网络通信。

### 创建代理

使用`rpc.InitProxy`创建代理：

```go
// 定义服务接口
type UserService struct {
    GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
    List    func(ctx context.Context, req *ListReq) (*ListResp, error)
}

// 创建代理
var userService UserService
err := rpc.InitProxy("localhost:8000", &userService)
if err != nil {
    log.Fatalf("Failed to init proxy: %v", err)
}
```

### 使用代理调用

```go
// 像调用本地方法一样调用远程服务
ctx := context.Background()
resp, err := userService.GetById(ctx, &GetByIdReq{Id: 123})
if err != nil {
    log.Fatalf("RPC call failed: %v", err)
}
fmt.Printf("User: %v\n", resp.User)
```

### 高级代理配置

代理支持多种高级配置：

```go
// 创建带负载均衡的代理
discoveryClient, _ := discovery.NewDiscovery(
    discovery.WithRegistry(etcdRegistry),
    discovery.WithBalancer(balancer.Random),
)
lb := discovery.NewLoadBalancer(discoveryClient, "user-service")

// 创建故障转移配置
failoverConfig := &failover.Config{
    MaxAttempts: 3,
    RetryInterval: time.Millisecond * 100,
}

// 初始化代理
var userService UserService
err := rpc.InitProxy(
    "localhost:8000", 
    &userService,
    rpc.WithLoadBalancer(lb),
    rpc.WithProxyFailover(failoverConfig),
    rpc.WithServiceName("UserService"),
)
```

### 代理工作原理

1. `InitProxy`通过反射分析服务接口结构
2. 对于每个函数字段，创建一个代理函数，替换原始字段
3. 代理函数负责：
    - 序列化请求参数
    - 选择合适的服务节点（如启用负载均衡）
    - 发送RPC请求
    - 处理故障转移（如配置）
    - 反序列化响应
    - 返回结果或错误

## 示例代码

### API客户端示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/fyerfyer/fyer-rpc/api"
    "github.com/fyerfyer/fyer-rpc/protocol"
)

// 请求和响应结构体
type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

func main() {
    // 创建客户端
    options := &api.ClientOptions{
        Address:       "localhost:8000",
        Timeout:       time.Second * 5,
        SerializeType: protocol.SerializationTypeJSON,
    }

    client, err := api.NewClient(options)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

    // 创建请求
    request := &HelloRequest{Name: "World"}
    response := &HelloResponse{}

    // 调用远程服务
    ctx := context.Background()
    err = client.Call(ctx, "GreeterService", "SayHello", request, response)
    if err != nil {
        log.Fatalf("RPC call failed: %v", err)
    }

    fmt.Printf("Response: %s\n", response.Message)
}
```

### 代理调用示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fyerfyer/fyer-rpc/rpc"
)

// 服务接口
type GreeterService struct {
    SayHello func(ctx context.Context, req *HelloRequest) (*HelloResponse, error)
}

// 请求和响应结构体
type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

func main() {
    // 创建代理
    var greeter GreeterService
    err := rpc.InitProxy("localhost:8000", &greeter)
    if err != nil {
        log.Fatalf("Failed to create proxy: %v", err)
    }

    // 调用远程方法
    ctx := context.Background()
    resp, err := greeter.SayHello(ctx, &HelloRequest{Name: "World"})
    if err != nil {
        log.Fatalf("RPC call failed: %v", err)
    }

    fmt.Printf("Response: %s\n", resp.Message)
}
```