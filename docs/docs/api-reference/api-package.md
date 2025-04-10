# `api` Package

`api`包提供了高级抽象的客户端和服务器接口，是使用 fyerrpc 框架最简单的方式。

### Server

`Server`接口定义了 RPC 服务器的核心功能：

```go
type Server interface {
    // Register 注册服务
    Register(service interface{}) error

    // Start 启动服务器
    Start() error

    // Stop 停止服务器
    Stop() error

    // Address 获取服务器监听地址
    Address() string
}
```

使用`NewServer`函数创建服务器实例：

```go
func NewServer(options *ServerOptions) Server
```

`ServerOptions`结构体用于配置服务器行为：

```go
type ServerOptions struct {
    Address        string   // 服务监听地址，默认":8000"
    SerializeType  uint8    // 序列化类型，默认JSON
    EnableRegistry bool     // 是否启用服务注册
    Registry       Registry // 注册中心实例
    RegistryAddrs  []string // 注册中心地址
    ServiceName    string   // 服务名称
    ServiceVersion string   // 服务版本
    Weight         int      // 服务权重
    Metadata       map[string]string // 服务元数据
}
```

### Client

`Client`接口定义了 RPC 客户端的核心功能：

```go
type Client interface {
    // Call 同步调用远程服务
    Call(ctx context.Context, service, method string, req interface{}, resp interface{}) error

    // Close 关闭客户端连接
    Close() error
}
```

使用`NewClient`函数创建客户端实例：

```go
func NewClient(options *ClientOptions) (Client, error)
```

`ClientOptions`结构体用于配置客户端行为：

```go
type ClientOptions struct {
    Address       string        // 服务器地址
    Timeout       time.Duration // 请求超时
    PoolSize      int           // 连接池大小
    SerializeType uint8         // 序列化类型
    Discovery     Discovery     // 服务发现实例
    ServiceName   string        // 服务名称
    ServiceVersion string       // 服务版本
}
```

### Service

`Service`接口允许开发者定义服务信息：

```go
type Service interface {
    // ServiceInfo 返回服务信息
    ServiceInfo() *ServiceInfo
}
```

`ServiceInfo`结构体包含服务的元数据：

```go
type ServiceInfo struct {
    Name        string            // 服务名称
    Version     string            // 服务版本
    Description string            // 服务描述
    Metadata    map[string]string // 服务元数据
}
```