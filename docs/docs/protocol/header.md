# header模块

`header.go`文件定义了fyer-rpc框架中协议头部的结构和相关常量。

## 协议头

头部结构的设计如下：

```
+-----------------------------------------------+
|  magic number   |  version    |  msg type     |
+-----------------------------------------------+
|  2 bytes        |  1 byte     |  1 byte       |
+-----------------------------------------------+
|  compress type  |  serial type|  message id   |
+-----------------------------------------------+
|  1 byte         |  1 byte     |  8 bytes      |
+-----------------------------------------------+
|  metadata size  |  payload size               |
+-----------------------------------------------+
|  4 bytes        |  4 bytes                    |
+-----------------------------------------------+
```

## 使用示例

```go
// 创建一个新的请求消息头部
requestHeader := protocol.Header{
    MagicNumber:       protocol.MagicNumber,
    Version:           1,
    MessageType:       protocol.TypeRequest,
    CompressType:      protocol.CompressTypeNone,
    SerializationType: protocol.SerializationTypeJSON,
    MessageID:         12345678,
    // MetadataSize和PayloadSize会在消息编码时设置
}

// 在完整的消息编码过程中使用
message := &protocol.Message{
    Header:   requestHeader,
    Metadata: &protocol.Metadata{
        ServiceName: "UserService",
        MethodName:  "GetUser",
    },
    Payload:  requestData,
}

// 然后使用protocol.EncodeMessage将消息编码并发送
```