# Protocol Package

protocol包定义了 RPC 通信协议，包括消息格式、序列化和编解码。

### Message

`Message`结构体定义了 RPC 消息的格式：

```go
type Message struct {
    Header   Header    // 消息头部
    Metadata *Metadata // 元数据
    Payload  []byte    // 消息体
}
```

### Header

`Header`结构体定义了消息头部：

```go
type Header struct {
    MagicNumber       uint16 // 魔数，用于校验报文
    Version           uint8  // 协议版本号
    MessageType       uint8  // 消息类型(请求/响应)
    CompressType      uint8  // 压缩类型
    SerializationType uint8  // 序列化类型
    MessageID         uint64 // 消息ID，用于多路复用
    MetadataSize      uint32 // 元数据长度
    PayloadSize       uint32 // 消息体长度
}
```

### Metadata

`Metadata`结构体包含 RPC 调用的元数据：

```go
type Metadata struct {
    ServiceName string            // 服务名称
    MethodName  string            // 方法名称
    Error       string            // 错误信息(仅响应消息使用)
    Extra       map[string]string // 额外的元数据，如trace_id等
}
```

### Protocol

Protocol接口定义了消息的编码和解码方法：

```go
type Protocol interface {
    EncodeMessage(message *Message, writer io.Writer) error
    DecodeMessage(reader io.Reader) (*Message, error)
}
```

### Codec

`codec`子包提供了不同的序列化实现：

```go
type Codec interface {
    // Encode 将对象序列化为字节数组
    Encode(v interface{}) ([]byte, error)

    // Decode 将字节数组反序列化为对象
    Decode(data []byte, v interface{}) error

    // Name 返回编解码器的名称
    Name() string
}
```

内置实现包括：
- `JsonCodec`: JSON 序列化
- `ProtobufCodec`: Protobuf 序列化