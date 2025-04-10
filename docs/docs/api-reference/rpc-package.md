# `rpc` Package

`rpc`包实现了 RPC 通信的核心逻辑，包括消息序列化、连接管理和服务注册等。

### Server

`Server`结构体实现了 RPC 服务器：

```go
type Server struct {
    services          map[string]*ServiceDesc
    serializationType uint8
}

// NewServer 创建新的RPC服务器
func NewServer() *Server

// RegisterService 注册服务实例
func (s *Server) RegisterService(service interface{}) error

// Start 启动服务器
func (s *Server) Start(address string) error
```

### Client

`Client`结构体实现了 RPC 客户端：

```go
type Client struct {
    conn            *Connection
    messageID       uint64
    failoverConfig  *failover.Config
    failoverHandler *failover.DefaultFailoverHandler
    enableFailover  bool
}

// NewClient 创建新的RPC客户端
func NewClient(address string, options ...ClientOption) (*Client, error)

// Call 调用远程服务
func (c *Client) Call(serviceName, methodName string, args interface{}) ([]byte, error)

// CallWithTimeout 带超时的RPC调用
func (c *Client) CallWithTimeout(ctx context.Context, serviceName, methodName string, args interface{}) ([]byte, error)
```

### Connection

`Connection`结构体封装了底层网络连接，提供协议层面的读写：

```go
type Connection struct {
    conn     net.Conn
    protocol protocol.Protocol
}

// Write 发送消息
func (c *Connection) Write(serviceName, methodName string, messageType uint8, serializationType uint8, messageID uint64, metadata *protocol.Metadata, payload []byte) error

// Read 接收消息
func (c *Connection) Read() (*protocol.Message, error)
```

### ConnPool

`ConnPool`结构体实现了连接池，用于管理和复用客户端连接：

```go
type ConnPool struct {
    address     string
    maxIdle     int
    idleTimeout time.Duration
    conns       chan *Client
}

// NewConnPool 创建连接池
func NewConnPool(address string, maxIdle int, idleTimeout time.Duration) *ConnPool

// Get 获取连接
func (p *ConnPool) Get() (*Client, error)

// Put 归还连接
func (p *ConnPool) Put(client *Client)
```

### Proxy

`Proxy`结构体实现了服务代理，用于创建客户端代理对象：

```go
// InitProxy 初始化服务代理
func InitProxy(address string, target interface{}, opts ...ProxyOption) error
```