# protocol模块

`protocol.go`文件是fyer-rpc框架的协议层核心实现，主要负责消息的编码与解码过程。该文件是整个协议栈的关键部分，直接负责网络传输数据的组装和解析。

## `Protocol`接口

该模块定义了协议编解码的核心接口，规定了协议实现类必须提供的功能：

```go
// Protocol 协议编解码接口
type Protocol interface {
    EncodeMessage(message *Message, writer io.Writer) error
    DecodeMessage(reader io.Reader) (*Message, error)
}
```

接口中：

* `EncodeMessage`：将消息结构编码为二进制数据并写入指定的输出流。
* `DecodeMessage`：从输入流中读取二进制数据并解码为消息结构。

## `DefaultProtocol`实现

`DefaultProtocol`是fyer-rpc的默认协议实现，具有如下特性：

* **字节序处理**：统一使用大端字节序（Big-Endian）进行二进制数据的编码和解码：

```go
if err := binary.Write(writer, binary.BigEndian, message.Header.MagicNumber); err != nil {
    return err
}
```

* **动态序列化支持**：协议支持根据消息头部中的序列化类型字段，动态选择不同的序列化方式：

```go
codec := GetCodecByType(message.Header.SerializationType)
if codec == nil {
    return ErrUnsupportedSerializer
}
```

## 使用示例

```go
// 创建协议实例
protocol := &protocol.DefaultProtocol{}

// 创建消息
message := &protocol.Message{
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
        },
    },
    Payload: []byte(`{"id": 1001}`),
}

// 编码消息
var buffer bytes.Buffer
err := protocol.EncodeMessage(message, &buffer)
if err != nil {
    log.Fatalf("failed to encode message: %v", err)
}

// 解码消息
decodedMessage, err := protocol.DecodeMessage(&buffer)
if err != nil {
    log.Fatalf("解码消息失败: %v", err)
}

// 使用解码后的消息
fmt.Printf("服务名: %s, 方法名: %s\n", 
    decodedMessage.Metadata.ServiceName, 
    decodedMessage.Metadata.MethodName)
```
