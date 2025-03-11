# etcd/registry模块

registry.go 文件实现了基于 etcd 的服务注册中心，是框架中 Registry 接口的具体实现。

## 1.设计思路

### `EtcdRegistry`组件

`EtcdRegistry`包含以下字段：

* `client`：etcd客户端
* locks：服务实例与租约 ID 的映射关系，使用并发安全的 sync.Map
* watchers：服务订阅的通道映射，两级结构：服务名称 -> 版本 -> 通知通道