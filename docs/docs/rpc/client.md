# client模块

`client.go`文件是fyer-rpc的客户端实现。


## 1.设计思路

### `Client`组件

`Client`包含以下字段：
- `conn`：与服务端的连接。
- `messageID`：消息ID，用于标识请求。
- 故障转移相关配置。

```go
type Client struct {
    conn            *Connection
    messageID       uint64
    failoverConfig  *failover.Config                 // 故障转移配置
    failoverHandler *failoverDefaultFailoverHandler  // 故障转移处理器
    enableFailover  bool                             // 是否启用故障转移
}
```

### `Call`方法

`Call`方法向指定的服务发送请求并等待响应，大致执行流程如下：

1. 构造方法元数据，序列化参数，然后发送请求。
2. 接收响应，解析出元数据与消息体。

具体的解析与构造过程在`protocol`目录中有详细介绍。

### `CallWithTimeout`方法

`CallWithTimeout`方法在`Call`的基础上开了一个`done`通道来监控方法调用的完成情况，当方法调用完后`done`通道会被关闭、同时收到停止信号。用`select`来判断先收到`done`的信号还是先收到`ctx.Done()`的信号。

### `CallWithFailover`方法

`CallWithFailover`的大致执行流程如下：

1. 使用`failoverHandler`执行调用，并获取返回的实例。
2. 如果调用成功，使用返回的实例执行真正的`Call`调用；否则返回报错。

## 2.使用示例

```go
type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

func main() {
    // 创建启用错误转移的客户端
    failoverConfig := &failover.Config{
        RetryTimes:      3,
        RetryInterval:   time.Second,
        TimeoutPerRetry: 2 * time.Second,
    }
    
    client, err := rpc.NewClient("localhost:8080", rpc.WithFailover(failoverConfig))
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    // 执行带有超时控制的rpc调用
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    request := &HelloRequest{Name: "World"}
    responseBytes, err := client.CallWithTimeout(ctx, "HelloService", "SayHello", request)
    if err != nil {
        fmt.Printf("RPC call failed: %v\n", err)
        return
    }
    
    // 解析响应
    var response HelloResponse
    if err := json.Unmarshal(responseBytes, &response); err != nil {
        fmt.Printf("Failed to parse response: %v\n", err)
        return
    }
    
    fmt.Println("Service response:", response.Message)
}
```

