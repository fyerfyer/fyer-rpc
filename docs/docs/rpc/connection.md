# connection模块

`connection.go` 文件是fyer-rpc的连接抽象层实现。它封装了底层的TCP连接，提供了基于fyer-rpc协议的消息读写能力。

## 设计思路

### `Connection`组件

`Connection`结构体是该模块的核心，它封装了原始网络连接和协议处理器：

```go
type Connection struct {
    conn     net.Conn
    protocol protocol.Protocol
}
```