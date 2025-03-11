# pool模块

`pool.go`文件是fyer-rpc框架的连接池功能实现，用于管理和复用客户端连接。

## 设计思路

### `ConnPool`组件

`ConnPool`包含了连接信息管理、空闲连接存储、并发控制等常见字段：

```go
type ConnPool struct {
    mu          sync.Mutex     // 互斥锁，保证并发安全
    address     string         // 服务器地址
    maxIdle     int            // 最大空闲连接数
    idleTimeout time.Duration  // 空闲连接超时时间
    conns       chan *Client   // 连接通道，存储空闲连接
}
```

### `Get`方法

`Get`方法尝试从连接池获取连接，如果连接池中有空闲连接的话就取出空闲连接使用，否则就创建一个连接。

### `Put`方法

`Put`方法将连接放回连接池中，供后续的rpc调用。如果连接池装空闲连接的通道已经关闭了，就直接关掉客户端、不再复用。

### `Close`方法

`Close`方法关闭连接池以及连接池通道内的所有空闲连接。
