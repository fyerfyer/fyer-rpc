# FyerRPC

FyerRPC是一个轻量级、高性能的Go语言RPC框架，提供了简单易用的API，同时具备服务发现、负载均衡、故障转移等企业级特性。

## 功能特性

- **简洁API**：易于集成和使用的客户端和服务端API
- **高性能**：针对高并发场景优化的连接池和协议设计
- **多种序列化**：支持JSON和Protobuf序列化
- **服务发现**：内置etcd服务注册与发现
- **负载均衡**：支持多种负载均衡策略（随机、轮询、最快响应）
- **故障转移**：自动故障检测和故障转移机制
- **熔断保护**：支持熔断器模式防止级联失败
- **分组路由**：基于标签和元数据的服务分组和路由
- **指标监控**：集成Prometheus的指标收集和监控

## 项目结构

```
fyerrpc/
├── api/              # 客户端和服务接口定义
├── cluster/          # 集群管理相关
│   ├── failover/     # 故障转移实现
│   ├── group/        # 分组路由实现
│   └── selector/     # 服务选择器
├── config/           # 配置管理
├── discovery/        # 服务发现
│   ├── balancer/     # 负载均衡
│   └── metrics/      # 指标收集
├── naming/           # 命名服务
├── protocol/         # 通信协议
│   └── codec/        # 编解码器
├── registry/         # 服务注册
│   └── etcd/         # etcd实现
├── rpc/              # RPC核心实现
└── utils/            # 工具类
```

## 安装

```bash
go get github.com/fyerfyer/fyer-rpc
```

## 快速开始

### 定义服务接口

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
// 方法的入参固定为*context.Context与*Request
// 方法的返回值固定为*Response与error

type UserService struct {
    GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}
```

### 实现服务

```go
type UserServiceImpl struct{}

// 注意：服务实现的命名规范为：服务名(UserService)+Impl
// 并且方法必须实现在结构体的接口上

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

### 创建服务端

这里仅展示通过`api`包创建服务端的方式，使用底层`rpc`包创建的方法参见docs文档。

```go
func main() {

    // 创建服务器
    server := api.NewServer(&api.ServerOptions{
        Address: ":8000", // 服务监听地址
        SerializeType: protocol.SerializationTypeJSON, // 使用JSON序列化
    })
	
    // 注册服务
    service := &service.UserService{}
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

### 创建客户端

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

## 高级功能

### 使用代理简化调用

FyerRPC提供了代理功能，用户可以选择使用代理、像调用本地方法一样调用远程方法，而不需要手动创建`Client`、调用`Call`方法、处理序列化/反序列化等，只需要在服务端实现接口并注册服务即可。

```go
// 初始化代理
var userService UserService 
err := rpc.InitProxy("localhost:8080", &userService)
if err != nil {
    panic(err)
}

// 直接调用，像本地方法一样
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
resp, err := userService.GetById(ctx, &GetByIdReq{Id: 123})
if err != nil {
    fmt.Printf("fail to call remote method: %v\n", err)
    return
}

// 直接使用响应（无需手动解析）
fmt.Printf("用户信息: ID=%d, Name=%s\n", resp.User.Id, resp.User.Name)
```

### 使用服务发现

```go
// 创建etcd注册中心
registry, err := etcd.New(
    etcd.WithEndpoints([]string{"localhost:2379"}),
    etcd.WithDialTimeout(time.Second*5),
)

// 创建服务发现
discovery := discovery.NewDiscovery(registry, time.Minute)

// 使用服务发现创建负载均衡器
balancer, err := discovery.NewLoadBalancer(
    "user-service", "1.0.0",
    balancer.Random,
)

// 创建使用负载均衡的客户端代理
var userService service.UserService
err = rpc.InitProxy("", &userService, 
    rpc.WithLoadBalancer(balancer),
)
```

### 配置故障转移

```go
// 配置故障转移
failoverConfig := &failover.Config{
    MaxRetries:    3,
    RetryInterval: 100 * time.Millisecond,
    RetryStrategy: "jittered",
}

// 创建客户端时启用故障转移
client, err := api.NewClient(&api.ClientOptions{
    Address: ":8080",
}, api.WithFailover(failoverConfig))
```

## 配置选项

FyerRPC提供了丰富的配置选项，可以通过config包下的各种配置类型进行详细控制：

- `config.ClientConfig`: 客户端配置
- `config.ServerConfig`: 服务端配置
- `cluster.failover.Config`: 故障转移配置
- `discovery.balancer.Config`: 负载均衡配置

## 许可证

项目使用MIT License。