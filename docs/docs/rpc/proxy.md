# proxy模块

`proxy.go`文件是fyer-rpc的动态代理实现，让用户不需要手动创建客户端、处理序列化/反序列化等细节。

## 1.设计思路

### `Proxy`组件

`Proxy`包含以下字段：

- `pool`：连接池，管理与服务器的连接。
- 负载均衡器与故障转移相关配置。

```go
type Proxy struct {
    pool           *ConnPool
    loadBalancer   *discovery.LoadBalancer // 负载均衡器
    enableFailover bool                    // 是否启用故障转移
    failoverConfig *failover.Config        // 故障转移配置
}
```

### `InitProxy`方法

`InitProxy`方法初始化了代理组件，大致执行流程如下：

1. 验证目标对象的正确性（是否为结构体指针）
2. 创建连接池，之后的代理请求都通过从连接池里面取客户端进行。
3. 使用`handleStructField`处理结构体字段的代理创建。

下面详细讲解`handleStructField`方法对结构体的处理：

### `handleStructField`方法`

`handleStructField`方法将结构体字段的方法对应的rpc函数篡改成实际的调用，也就是说，对于下面的结构：

```go
type UserService struct {
    GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}
```

用户在使用代理对其中的方法进行调用时，由于代理将`GetById`篡改成了实际的rpc调用，用户就不需要处理这些细节了。

这一过程的具体实现逻辑如下：

1. 使用`reflect.MakeFunc()`创建和原有结构体内方法字段（即`structField.Type`）相同的篡改方法。
2. 从连接池中取一个客户端来执行rpc调用。
3. 调用客户端的`CallWithTimeout`方法执行rpc调用。
4. 返回调用结果。

在实现完篡改方法后，把原有结构体方法字段的值替换为这个篡改方法即可。

### 2. 使用示例

```go
// 定义服务结构体
type HelloService struct {
    SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error)
}

// 请求和响应结构
type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

func main() {
    // 创建服务代理
    var helloService HelloService
    err := rpc.InitProxy(
        "localhost:8080", 
        &helloService,
    )
    if err != nil {
        panic(err)
    }
    
    // 像调用本地方法一样使用远程服务
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    response, err := helloService.SayHello(ctx, &HelloRequest{Name: "fyerfyer"})
    if err != nil {
        fmt.Printf("failed to call rpc method: %v\n", err)
        return
    }
    
    fmt.Println("service response: ", response.Message)
}
```