# Server

服务器是fyerrpc框架的核心组件之一，负责接收和处理客户端的RPC请求，调用相应的服务方法，并将结果返回给客户端。

## 基础服务器

### Server接口

fyerrpc框架通过`api.Server`接口定义了服务器的基本行为：

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

### 创建服务器实例

创建一个基本的fyerrpc服务器非常简单：

```go
import "github.com/fyerfyer/fyer-rpc/api"

// 创建服务器配置
options := &api.ServerOptions{
    Address:        ":8000",                          // 服务监听地址
    SerializeType:  protocol.SerializationTypeJSON,   // 使用JSON序列化
    EnableRegistry: false,                           // 不启用服务注册
}

// 创建服务器
server := api.NewServer(options)
```

如果不提供配置项，服务器将使用默认配置：

```go
// 使用默认配置创建服务器
server := api.NewServer(nil) // 等同于使用DefaultServerOptions
```

### 注册服务

在启动服务器前，需要注册服务实现：

```go
// 定义服务实现
type GreeterServiceImpl struct{}

func (s *GreeterServiceImpl) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
    return &HelloResponse{
        Message: fmt.Sprintf("Hello, %s!", req.Name),
    }, nil
}

// 注册服务
err := server.Register(&GreeterService{})
if err != nil {
    log.Fatalf("Failed to register service: %v", err)
}
```

注册服务时，框架会自动提取服务名称和方法信息，无需额外配置。

### 启动和停止服务器

启动服务器：

```go
// 启动服务器（非阻塞）
if err := server.Start(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}
log.Println("Server started at", server.Address())
```

优雅停止服务器：

```go
// 等待终止信号
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
<-sigCh

// 停止服务器
log.Println("Shutting down server...")
server.Stop()
```

## 底层服务器实现

除了高级的`api.Server`接口，fyerrpc还提供了更底层的`rpc.Server`实现，为需要更多控制和定制的场景提供支持。

### 创建底层服务器

```go
import "github.com/fyerfyer/fyer-rpc/rpc"

// 创建服务器实例
server := rpc.NewServer()

// 设置序列化类型（可选）
server.SetSerializationType(protocol.SerializationTypeJSON)
```

### 注册服务

底层服务器使用`RegisterService`方法注册服务：

```go
// 注册服务
err := server.RegisterService(&GreeterService{})
if err != nil {
    log.Fatalf("Failed to register service: %v", err)
}
```

与`api.Server`相同，底层服务器会通过反射提取服务结构体信息。服务名使用结构体的名称（如果有`Impl`后缀会被自动移除）。

### 启动底层服务器

底层服务器使用`Start`方法启动，需要直接指定监听地址：

```go
// 启动服务器（阻塞）
if err := server.Start(":8000"); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}
```

> 注意，与`api.Server`不同，`rpc.Server`的`Start`方法是阻塞的，通常应该在一个单独的goroutine中调用：

```go
go func() {
    if err := server.Start(":8000"); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}()
```

### 底层服务器处理流程

底层服务器的请求处理流程如下：

1. 接受客户端连接
2. 为每个连接创建新的goroutine
3. 读取RPC请求消息
4. 查找对应的服务和方法
5. 解码请求参数
6. 调用服务方法
7. 编码响应结果
8. 将响应发送回客户端

底层服务器处理错误的方式也更加直接，如找不到服务或方法时会立即将错误响应返回给客户端。

## 服务器配置

fyerrpc提供了丰富的服务器配置选项，分为通用配置和服务器特定配置两部分。

### ServerConfig

`config.ServerConfig`包含了详细的服务器配置项：

```go
type ServerConfig struct {
    *CommonConfig // 继承通用配置

    // 网络相关配置
    Address         string        // 服务监听地址
    MaxConnections  int           // 最大连接数
    ReadTimeout     time.Duration // 读取超时
    WriteTimeout    time.Duration // 写入超时
    ShutdownTimeout time.Duration // 优雅关闭超时
    MaxHeaderBytes  int           // 最大请求头大小

    // 注册中心配置
    RegisterTTL      int64             // 服务注册租约时间（秒）
    RegisterInterval time.Duration     // 服务注册间隔
    EnableRegistry   bool              // 是否启用服务注册
    ServiceName      string            // 服务名称
    ServiceVersion   string            // 服务版本
    ServiceWeight    int               // 服务权重
    Metadata         map[string]string // 服务元数据

    // 处理器配置
    WorkerPoolSize   int           // 工作线程池大小
    MaxRequestSize   int           // 最大请求大小(字节)
    MaxConcurrent    int           // 最大并发请求数
    SlowRequestTime  time.Duration // 慢请求阈值
    EnableAccessLog  bool          // 是否启用访问日志
    EnableMetricsLog bool          // 是否启用指标日志
    MetricsLogPeriod time.Duration // 指标日志周期
}
```

### 默认配置

fyerrpc为ServerConfig提供了合理的默认值：

```go
var DefaultServerConfig = &ServerConfig{
    CommonConfig: DefaultCommonConfig,

    // 网络相关默认配置
    Address:         ":8000",
    MaxConnections:  1000,
    ReadTimeout:     time.Second * 30,
    WriteTimeout:    time.Second * 30,
    ShutdownTimeout: time.Second * 10,
    MaxHeaderBytes:  1 << 20, // 1MB

    // 注册中心默认配置
    RegisterTTL:      30,
    RegisterInterval: time.Second * 10,
    EnableRegistry:   false,
    ServiceName:      "",
    ServiceVersion:   "1.0.0",
    ServiceWeight:    100,
    Metadata:         make(map[string]string),

    // 处理器默认配置
    WorkerPoolSize:   runtime.NumCPU() * 2,
    MaxRequestSize:   4 << 20, // 4MB
    MaxConcurrent:    100,
    SlowRequestTime:  time.Second * 1,
    EnableAccessLog:  true,
    EnableMetricsLog: false,
    MetricsLogPeriod: time.Minute,
}
```

### 配置选项函数

fyerrpc采用选项模式，通过函数链式调用进行配置：

```go
// 创建服务器配置
config := config.NewServerConfig(
    // 设置网络相关配置
    config.WithAddress(":9000"),
    config.WithNetworkConfig(2000, time.Second*60, time.Second*60),
    
    // 设置服务信息
    config.WithServiceInfo("myservice", "2.0.0", 150),
    
    // 设置注册中心配置
    config.WithRegistryConfig(true, 60, time.Second*20),
    
    // 设置工作池大小
    config.WithWorkerPoolSize(runtime.NumCPU()*4),
)

// 应用通用配置
config.Apply(
    config.WithSerialization(config.SerializationJSON),
    config.WithLogLevel(utils.InfoLevel),
)

// 初始化配置
config.Init()
```

### 通用配置

`CommonConfig`包含了服务器和客户端共享的配置项：

```go
type CommonConfig struct {
    // 日志配置
    LogLevel     utils.LogLevel // 日志级别
    LogOutput    io.Writer      // 日志输出
    EnabledDebug bool           // 是否启用调试日志

    // 协议配置
    SerializationType SerializationType // 序列化类型
    CompressType      CompressType      // 压缩类型
    ProtocolVersion   uint8             // 协议版本

    // 超时配置
    DialTimeout    time.Duration // 连接超时
    RequestTimeout time.Duration // 请求超时

    // 重试配置
    MaxRetries     int           // 最大重试次数
    RetryInterval  time.Duration // 重试间隔
    RetryableError []string      // 可重试的错误类型

    // 注册中心配置
    RegistryType     string   // 注册中心类型（etcd, consul, etc.）
    RegistryEndpoint []string // 注册中心地址

    // 指标配置
    EnableMetrics   bool          // 是否启用指标收集
    MetricsInterval time.Duration // 指标收集间隔

    // 元数据
    Metadata map[string]string
}
```

## 生命周期管理

### 服务器启动流程

fyerrpc服务器的启动流程如下：

1. **创建服务器实例**：通过`api.NewServer()`或`rpc.NewServer()`创建服务器实例
2. **注册服务**：调用`server.Register()`或`server.RegisterService()`注册服务实现
3. **启动服务**：调用`server.Start()`启动网络监听
4. **处理连接**：接受客户端连接并启动处理协程
5. **读取请求**：从连接中读取RPC请求
6. **查找服务**：根据请求中的服务名和方法名查找对应的服务实现
7. **反序列化参数**：将请求参数反序列化为目标类型
8. **调用方法**：通过反射调用服务方法
9. **返回响应**：将方法执行结果序列化后发送给客户端

### 服务注册与发现

当启用服务注册功能时(`EnableRegistry=true`)，服务器启动时会：

1. 根据配置连接注册中心（默认支持etcd）
2. 将服务信息注册到注册中心
3. 定期向注册中心发送心跳，续约服务租约
4. 服务停止时，从注册中心注销服务信息

```go
// 启用服务注册
options := &api.ServerOptions{
    Address:        ":8000",
    EnableRegistry: true,
    RegistryAddrs:  []string{"localhost:2379"},
    ServiceName:    "greeter-service",
    ServiceVersion: "1.0.0",
}
```

### 优雅关闭

服务器的优雅关闭流程：

1. 接收到停止信号
2. 停止接受新的连接
3. 等待正在处理的请求完成，最长等待时间由`ShutdownTimeout`控制
4. 关闭所有连接
5. 如果启用了服务注册，从注册中心注销服务
6. 释放资源

```go
// 设置优雅关闭超时
config := config.NewServerConfig(
    config.WithShutdownTimeout(time.Second * 30),
)
```

## 示例代码

### API服务器使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/fyerfyer/fyer-rpc/api"
    "github.com/fyerfyer/fyer-rpc/protocol"
    "github.com/fyerfyer/fyer-rpc/utils"
)

// 定义请求和响应结构体
type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

// 定义服务实现
type GreeterService struct{}

func (s *GreeterService) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
    return &HelloResponse{
        Message: fmt.Sprintf("Hello, %s!", req.Name),
    }, nil
}

func main() {
    // 配置日志
    utils.SetDefaultLogger(utils.NewLogger(utils.InfoLevel, os.Stdout))

    // 创建服务器选项
    options := &api.ServerOptions{
        Address:        ":8000",
        SerializeType:  protocol.SerializationTypeJSON,
        EnableRegistry: false,
    }

    // 创建服务器
    server := api.NewServer(options)

    // 注册服务
    err := server.Register(&GreeterService{})
    if err != nil {
        log.Fatalf("Failed to register service: %v", err)
    }

    // 启动服务器
    log.Println("Starting server on", options.Address)
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

### 底层服务器示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/fyerfyer/fyer-rpc/protocol"
    "github.com/fyerfyer/fyer-rpc/rpc"
    "github.com/fyerfyer/fyer-rpc/utils"
)

// 定义请求和响应结构体
type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

// 定义服务实现
type GreeterService struct{}

func (s *GreeterService) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
    return &HelloResponse{
        Message: fmt.Sprintf("Hello, %s!", req.Name),
    }, nil
}

func main() {
    // 配置日志
    utils.SetDefaultLogger(utils.NewLogger(utils.InfoLevel, os.Stdout))

    // 创建底层服务器
    server := rpc.NewServer()
    
    // 设置序列化类型
    server.SetSerializationType(protocol.SerializationTypeJSON)

    // 注册服务
    err := server.RegisterService(&GreeterService{})
    if err != nil {
        log.Fatalf("Failed to register service: %v", err)
    }

    // 启动服务器（在后台运行）
    address := ":8000"
    log.Println("Starting RPC server on", address)
    go func() {
        if err := server.Start(address); err != nil {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // 等待终止信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh

    log.Println("Shutting down server...")
    // 底层服务器当前不支持直接停止，但可以通过关闭监听器实现
}
```