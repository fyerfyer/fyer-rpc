# message模块

`message.go`文件定义了fyer-rpc框架中的消息结构体系，包括完整消息结构、元数据结构以及相关的接口实现。它是协议层的核心组件，负责描述rpc通信过程中请求和响应消息的格式与内容。

## 消息格式

消息格式的设计如下：

```
// Message 完整的消息结构
// 消息格式:
// +------------------+
// |     Header      |  消息头部（固定22字节）
// +------------------+
// |    Metadata     |  元数据(可变长度)，包含服务名、方法名等信息
// +------------------+
// |    Payload      |  消息体(可变长度)，包含请求参数或响应结果
// +------------------+
type Message struct {
    Header   Header    // 消息头部
    Metadata *Metadata // 元数据
    Payload  []byte    // 消息体
}
```

## 元数据

元数据结构的设计如下：

```go
// Metadata 元数据结构
// 包含RPC调用的必要信息和可选的链路追踪信息
type Metadata struct {
    ServiceName string            // 服务名称
    MethodName  string            // 方法名称
    Error       string            // 错误信息(仅响应消息使用)
    Extra       map[string]string // 额外的元数据，如trace_id等
}
```

元数据中的`Extra`字段提供了一个灵活的扩展机制，允许在不修改核心结构的情况下添加额外的上下文信息。用户可以通过使用该字段实现以下的功能：

* **分布式追踪**：可以在`Extra`中传递`trace_id`等追踪信息
* **认证授权**：可以传递身份令牌或权限信息
* **自定义控制**：可以传递调用超时、重试策略等控制参数

## 使用示例

```go
// 创建请求消息
requestMsg := &protocol.Message{
    Header: protocol.Header{
        MagicNumber:       protocol.MagicNumber,
        Version:           1,
        MessageType:       protocol.TypeRequest,
        CompressType:      protocol.CompressTypeNone,
        SerializationType: protocol.SerializationTypeJSON,
        MessageID:         12345678,
    },
    Metadata: &protocol.Metadata{
        ServiceName: "UserService",
        MethodName:  "GetUser",
        Extra: map[string]string{
            "trace_id": "abc-123-xyz",
            "timeout":  "5s",
        },
    },
    Payload: []byte(`{"id": 1001}`),
}

// 创建响应消息
responseMsg := &protocol.Message{
    Header: protocol.Header{
        MagicNumber:       protocol.MagicNumber,
        Version:           1,
        MessageType:       protocol.TypeResponse,
        CompressType:      protocol.CompressTypeNone,
        SerializationType: protocol.SerializationTypeJSON,
        MessageID:         12345678, // 与请求消息ID相同
    },
    Metadata: &protocol.Metadata{
        ServiceName: "UserService",
        MethodName:  "GetUser",
        // 如果调用成功，Error为空
        // 如果调用失败，设置错误信息
        Error: "",
        Extra: map[string]string{
            "trace_id": "abc-123-xyz",
        },
    },
    Payload: []byte(`{"id": 1001, "name": "张三", "age": 30}`),
}

// 处理响应消息中的错误
if responseMsg.Error() != "" {
    // 处理错误
    fmt.Printf("RPC调用失败: %s\n", responseMsg.Error())
    return
}

// 处理正常响应数据
// ...解析Payload...
```