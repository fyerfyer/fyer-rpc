# Service Registry & Discovery
    
服务注册与发现是分布式系统中的关键组件，它允许服务动态地注册自己的位置，并让其他服务能够发现和调用它们。

## 基础概念

### 服务注册

服务注册是指服务实例在启动时将自己的地址和元数据信息注册到注册中心，使得其他服务能够发现它。

### 服务发现

服务发现是指客户端从注册中心获取可用服务实例列表，并根据负载均衡策略选择合适的实例进行调用。

### 服务实例

fyerrpc使用`naming.Instance`结构来表示一个服务实例：

```go
type Instance struct {
    ID       string            // 实例唯一ID
    Service  string            // 服务名称
    Version  string            // 服务版本
    Address  string            // 服务地址(host:port)
    Status   Status            // 服务状态
    Weight   int               // 服务权重，用于负载均衡
    Metadata map[string]string // 服务元数据
}
```

## 注册中心接口

fyerrpc定义了通用的注册中心接口，所有注册中心实现必须遵循这个接口：

```go
type Registry interface {
    // Register 注册服务实例
    Register(ctx context.Context, service *naming.Instance) error

    // Deregister 注销服务实例
    Deregister(ctx context.Context, service *naming.Instance) error

    // Subscribe 订阅服务变更
    Subscribe(ctx context.Context, service string, version string) (<-chan []*naming.Instance, error)

    // Unsubscribe 取消订阅服务变更
    Unsubscribe(ctx context.Context, service string, version string) error

    // ListServices 获取服务实例列表
    ListServices(ctx context.Context, service string, version string) ([]*naming.Instance, error)

    // UpdateService 服务心跳
    UpdateService(ctx context.Context, service *naming.Instance) error

    // Close 关闭注册中心连接
    Close() error
}
```

## etcd 注册中心

fyerrpc默认提供了基于etcd的注册中心实现，etcd是一个分布式键值存储系统，适合用作服务注册中心。

### 创建etcd注册中心

使用`etcd.New`函数创建etcd注册中心：

```go
import (
    "github.com/fyerfyer/fyer-rpc/registry/etcd"
)

registry, err := etcd.New(
    etcd.WithEndpoints([]string{"localhost:2379"}),
    etcd.WithDialTimeout(time.Second*5),
    etcd.WithTTL(30), // 服务租约时间，单位秒
)
if err != nil {
    log.Fatalf("Failed to create etcd registry: %v", err)
}
defer registry.Close()
```

### 配置选项

etcd注册中心支持多种配置选项：

```go
// 基本配置
registry, err := etcd.New(
    etcd.WithEndpoints([]string{"localhost:2379", "localhost:2380"}), // etcd集群地址
    etcd.WithDialTimeout(time.Second*5),                             // 连接超时时间
    etcd.WithTTL(30),                                                // 服务租约时间(秒)
    etcd.WithNamespace("myapp"),                                     // 服务命名空间
    etcd.WithMaxRetry(5),                                            // 最大重试次数
    etcd.WithRetryInterval(time.Second),                             // 重试间隔
)

// 安全配置
registry, err := etcd.New(
    etcd.WithUsername("etcd-user"),                                   // etcd用户名
    etcd.WithPassword("etcd-password"),                               // etcd密码
    etcd.WithTLSConfig("cert.pem", "key.pem", "ca.pem"),             // TLS配置
)

// 高级配置
registry, err := etcd.New(
    etcd.WithAutoSyncInterval(time.Minute*5),                         // 自动同步成员列表的间隔
    etcd.WithDialKeepAlive(time.Second*30),                           // KeepAlive探测间隔
    etcd.WithCache(true, time.Minute),                                // 启用本地缓存
)
```

### 注册服务

服务端启动时，需要将服务实例注册到注册中心：

```go
import (
    "github.com/fyerfyer/fyer-rpc/naming"
    "github.com/google/uuid"
)

// 创建服务实例
instance := &naming.Instance{
    ID:      uuid.New().String(),                  // 实例唯一ID
    Service: "user-service",                       // 服务名称
    Version: "1.0.0",                              // 服务版本
    Address: "192.168.1.100:8000",                 // 服务地址
    Status:  naming.StatusEnabled,                 // 服务状态
    Weight:  100,                                  // 服务权重
    Metadata: map[string]string{                   // 服务元数据
        "region": "cn-shanghai",
        "zone":   "cn-shanghai-a",
    },
}

// 注册服务
ctx := context.Background()
err = registry.Register(ctx, instance)
if err != nil {
    log.Fatalf("Failed to register service: %v", err)
}

// 程序退出时注销服务
defer registry.Deregister(ctx, instance)
```

### 服务发现

客户端需要发现和订阅服务实例：

```go
// 获取服务实例列表
ctx := context.Background()
instances, err := registry.ListServices(ctx, "user-service", "1.0.0")
if err != nil {
    log.Fatalf("Failed to list services: %v", err)
}

// 处理服务实例
for _, instance := range instances {
    fmt.Printf("Found instance: %s at %s\n", instance.ID, instance.Address)
}
```

### 订阅服务变更

客户端可以订阅服务实例的变更：

```go
// 订阅服务变更
ctx := context.Background()
watchChan, err := registry.Subscribe(ctx, "user-service", "1.0.0")
if err != nil {
    log.Fatalf("Failed to subscribe service: %v", err)
}

// 处理服务变更
go func() {
    for {
        select {
        case instances, ok := <-watchChan:
            if !ok {
                log.Println("Watch channel closed")
                return
            }
            log.Printf("Service instances updated, count: %d\n", len(instances))
            // 处理更新后的实例列表...
        }
    }
}()

// 取消订阅
defer registry.Unsubscribe(ctx, "user-service", "1.0.0")
```

### 服务心跳

服务实例需要定期向注册中心发送心跳，表明自己仍然在线：

```go
// 启动心跳协程
go func() {
    ticker := time.NewTicker(time.Second * 10)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            err := registry.UpdateService(ctx, instance)
            if err != nil {
                log.Printf("Failed to update service: %v", err)
            }
        case <-ctx.Done():
            return
        }
    }
}()
```

> 注意：在etcd注册中心实现中，`Register`方法会自动启动心跳维护，无需手动发送心跳。

### etcd注册中心实现细节

fyerrpc的etcd注册中心使用以下设计：

1. **服务键格式**：`/fyerrpc/services/{service}/{version}/{id}`

2. **租约机制**：每个服务实例都绑定到etcd的租约，如果服务无法保持心跳，租约过期后服务实例会被自动删除

3. **监听机制**：使用etcd的Watch API来监听键前缀的变化，当服务实例发生变化时通知订阅者

4. **实例序列化**：服务实例信息使用JSON格式存储在etcd中

#### 服务注册流程

1. 创建租约
2. 将服务信息绑定到租约
3. 启动租约续期协程
4. 存储租约ID以便后续操作

#### 服务发现流程

1. 使用前缀查询获取所有匹配的服务实例
2. 解析每个实例的JSON数据
3. 返回实例列表

#### 服务订阅流程

1. 创建用于发送更新的通道
2. 获取当前服务列表作为初始数据
3. 启动后台监听goroutine
4. 当服务发生变化时，重新获取完整列表并发送到通道

## 自定义注册中心

fyerrpc支持自定义注册中心，只需实现`registry.Registry`接口即可。

### 实现注册中心接口

要创建自定义注册中心，需要实现以下方法：

```go
type MyRegistry struct {
    // 你的注册中心字段
}

// Register 注册服务实例
func (r *MyRegistry) Register(ctx context.Context, service *naming.Instance) error {
    // 实现注册逻辑
}

// Deregister 注销服务实例
func (r *MyRegistry) Deregister(ctx context.Context, service *naming.Instance) error {
    // 实现注销逻辑
}

// Subscribe 订阅服务变更
func (r *MyRegistry) Subscribe(ctx context.Context, service string, version string) (<-chan []*naming.Instance, error) {
    // 实现订阅逻辑
}

// Unsubscribe 取消订阅服务变更
func (r *MyRegistry) Unsubscribe(ctx context.Context, service string, version string) error {
    // 实现取消订阅逻辑
}

// ListServices 获取服务实例列表
func (r *MyRegistry) ListServices(ctx context.Context, service string, version string) ([]*naming.Instance, error) {
    // 实现列表获取逻辑
}

// UpdateService 服务心跳
func (r *MyRegistry) UpdateService(ctx context.Context, service *naming.Instance) error {
    // 实现心跳更新逻辑
}

// Close 关闭注册中心连接
func (r *MyRegistry) Close() error {
    // 实现清理逻辑
}
```

### 示例：基于内存的注册中心

下面是一个简单的基于内存的注册中心实现示例：

```go
package memory

import (
    "context"
    "sync"
    "time"

    "github.com/fyerfyer/fyer-rpc/naming"
    "github.com/fyerfyer/fyer-rpc/registry"
)

// MemoryRegistry 基于内存的注册中心实现
type MemoryRegistry struct {
    services   map[string]map[string]*naming.Instance     // service -> id -> instance
    subscribers map[string]map[string][]chan []*naming.Instance // service -> version -> channels
    mu         sync.RWMutex
}

// New 创建内存注册中心
func New() *MemoryRegistry {
    return &MemoryRegistry{
        services:    make(map[string]map[string]*naming.Instance),
        subscribers: make(map[string]map[string][]chan []*naming.Instance),
    }
}

// Register 注册服务实例
func (r *MemoryRegistry) Register(ctx context.Context, service *naming.Instance) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    key := service.Service
    if _, ok := r.services[key]; !ok {
        r.services[key] = make(map[string]*naming.Instance)
    }
    r.services[key][service.ID] = service

    // 通知订阅者
    r.notifySubscribersLocked(service.Service, service.Version)
    return nil
}

// Deregister 注销服务实例
func (r *MemoryRegistry) Deregister(ctx context.Context, service *naming.Instance) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    key := service.Service
    if instances, ok := r.services[key]; ok {
        delete(instances, service.ID)
        if len(instances) == 0 {
            delete(r.services, key)
        }
    }

    // 通知订阅者
    r.notifySubscribersLocked(service.Service, service.Version)
    return nil
}

// 通知订阅者（内部方法，需要在持有锁的情况下调用）
func (r *MemoryRegistry) notifySubscribersLocked(service, version string) {
    instances := r.getServicesLocked(service, version)

    if versionMap, ok := r.subscribers[service]; ok {
        if channels, ok := versionMap[version]; ok {
            for _, ch := range channels {
                // 非阻塞发送
                select {
                case ch <- instances:
                default:
                    // 通道已满，跳过
                }
            }
        }
    }
}

// getServicesLocked 获取服务实例（内部方法，需要在持有锁的情况下调用）
func (r *MemoryRegistry) getServicesLocked(service, version string) []*naming.Instance {
    var instances []*naming.Instance
    if svcMap, ok := r.services[service]; ok {
        for _, instance := range svcMap {
            if instance.Version == version {
                instances = append(instances, instance)
            }
        }
    }
    return instances
}

// Subscribe 订阅服务变更
func (r *MemoryRegistry) Subscribe(ctx context.Context, service string, version string) (<-chan []*naming.Instance, error) {
    r.mu.Lock()
    defer r.mu.Unlock()

    // 创建通道
    ch := make(chan []*naming.Instance, 10)

    // 初始化订阅映射
    if _, ok := r.subscribers[service]; !ok {
        r.subscribers[service] = make(map[string][]chan []*naming.Instance)
    }
    if _, ok := r.subscribers[service][version]; !ok {
        r.subscribers[service][version] = make([]chan []*naming.Instance, 0)
    }

    // 添加到订阅者列表
    r.subscribers[service][version] = append(r.subscribers[service][version], ch)

    // 发送当前实例列表作为初始数据
    instances := r.getServicesLocked(service, version)
    go func() {
        ch <- instances
    }()

    return ch, nil
}

// Unsubscribe 取消订阅服务变更
func (r *MemoryRegistry) Unsubscribe(ctx context.Context, service string, version string) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if versionMap, ok := r.subscribers[service]; ok {
        if channels, ok := versionMap[version]; ok {
            // 关闭所有通道
            for _, ch := range channels {
                close(ch)
            }
            // 删除订阅记录
            delete(versionMap, version)
            if len(versionMap) == 0 {
                delete(r.subscribers, service)
            }
        }
    }
    return nil
}

// ListServices 获取服务实例列表
func (r *MemoryRegistry) ListServices(ctx context.Context, service string, version string) ([]*naming.Instance, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.getServicesLocked(service, version), nil
}

// UpdateService 更新服务
func (r *MemoryRegistry) UpdateService(ctx context.Context, service *naming.Instance) error {
    return r.Register(ctx, service) // 简化实现，直接复用Register
}

// Close 关闭注册中心
func (r *MemoryRegistry) Close() error {
    r.mu.Lock()
    defer r.mu.Unlock()

    // 关闭所有订阅通道
    for _, versionMap := range r.subscribers {
        for _, channels := range versionMap {
            for _, ch := range channels {
                close(ch)
            }
        }
    }

    // 清空数据
    r.services = make(map[string]map[string]*naming.Instance)
    r.subscribers = make(map[string]map[string][]chan []*naming.Instance)

    return nil
}
```

## 在fyerrpc中使用注册中心

### 服务端集成

在服务端，可以通过配置启用服务注册：

```go
import (
    "github.com/fyerfyer/fyer-rpc/api"
    "github.com/fyerfyer/fyer-rpc/registry/etcd"
)

// 创建etcd注册中心
registry, err := etcd.New(
    etcd.WithEndpoints([]string{"localhost:2379"}),
)
if err != nil {
    log.Fatalf("Failed to create registry: %v", err)
}

// 创建服务器配置
options := &api.ServerOptions{
    Address:        ":8000",
    EnableRegistry: true,
    Registry:       registry,
    ServiceName:    "greeter-service",
    ServiceVersion: "1.0.0",
    Weight:         100,
    Metadata: map[string]string{
        "region": "cn-shanghai",
    },
}

// 创建服务器
server := api.NewServer(options)

// 注册服务
err = server.Register(&GreeterService{})
if err != nil {
    log.Fatalf("Failed to register service: %v", err)
}

// 启动服务器 (注册会自动发生)
if err := server.Start(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}
```

### 客户端集成

在客户端，可以使用服务发现来自动查找服务：

```go
import (
    "github.com/fyerfyer/fyer-rpc/api"
    "github.com/fyerfyer/fyer-rpc/discovery"
    "github.com/fyerfyer/fyer-rpc/discovery/balancer"
    "github.com/fyerfyer/fyer-rpc/registry/etcd"
)

// 创建etcd注册中心
registry, err := etcd.New(
    etcd.WithEndpoints([]string{"localhost:2379"}),
)
if err != nil {
    log.Fatalf("Failed to create registry: %v", err)
}

// 创建服务发现
discoveryClient, err := discovery.NewDiscovery(
    discovery.WithRegistry(registry),
    discovery.WithBalancer(balancer.Random),
)
if err != nil {
    log.Fatalf("Failed to create discovery: %v", err)
}

// 创建客户端配置
options := &api.ClientOptions{
    Discovery:     discoveryClient,
    ServiceName:   "greeter-service",
    ServiceVersion: "1.0.0",
    Timeout:       time.Second * 5,
}

// 创建客户端
client, err := api.NewClient(options)
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

// 调用服务 (会自动选择一个可用实例)
request := &HelloRequest{Name: "World"}
response := &HelloResponse{}
err = client.Call(context.Background(), "GreeterService", "SayHello", request, response)
```

