# naming

`naming`模块是fyer-rpc框架中负责服务命名和服务实例管理的核心组件。它定义了服务实例的数据结构和相关操作，为服务注册与发现提供基础支持

## 1.核心功能

`naming`模块实现了如下的功能：

* 定义服务实例的数据结构
* 提供服务键的生成和解析工具
* 管理服务实例的状态

## 2.模块设计

`naming`模块采用层次化的命名空间设计，使用路径格式来唯一标识服务实例。这种设计与常见的注册中心（etcd、ZooKeeper等）的数据组织方式保持一致，便于实现服务注册与发现功能。

### 服务键设计

服务键采用层次化路径格式：`/fyerrpc/services/{service}/{version}/{instance_id}`

* `/fyerrpc` - 框架根命名空间
* `/services` - 服务类型命名空间
* `{service}` - 具体服务名称
* `{version}` - 服务版本
* `{instance_id}` - 实例唯一标识

### 服务实例结构

`Instance`结构体包含了服务实例的完整信息：

```go
type Instance struct {
    ID        string            `json:"id"`         // 实例唯一标识
    Service   string            `json:"service"`    // 服务名称
    Version   string            `json:"version"`    // 服务版本
    Address   string            `json:"address"`    // 服务地址
    Metadata  map[string]string `json:"metadata"`   // 元数据
    Status    uint8             `json:"status"`     // 服务状态
    UpdatedAt int64             `json:"updated_at"` // 更新时间
}
```

### 服务状态管理
   服务状态管理模块定义了两种服务状态：

```go
const (
    StatusEnabled  = uint8(1) // 服务可用
    StatusDisabled = uint8(0) // 服务不可用
)
```

## 3.使用示例

```go
// 创建服务实例
instance := &naming.Instance{
    ID:        "instance-001",
    Service:   "user-service",
    Version:   "v1.0",
    Address:   "127.0.0.1:8080",
    Metadata:  map[string]string{"region": "cn-east"},
    Status:    naming.StatusEnabled,
    UpdatedAt: time.Now().Unix(),
}

// 构建服务键
serviceKey := naming.BuildServiceKey(instance.Service, instance.Version, instance.ID)
// 结果: /fyerrpc/services/user-service/v1.0/instance-001

// 解析服务键
service, version, instanceID, err := naming.ParseServiceKey(serviceKey)
if err != nil {
    // 处理错误
}
fmt.Printf("service: %s, version: %s, instance ID: %s\n", service, version, instanceID)
```