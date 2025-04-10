# Protocol

fyerrpc框架设计了自己的二进制协议来保证高效、可靠的RPC通信。

## 消息格式

fyerrpc采用简单高效的二进制格式，一个完整的RPC消息由三部分组成：协议头(Header)、元数据(Metadata)和消息体(Payload)。

### 整体结构

```
+------------------+
|     Header      |  消息头部（固定22字节）
+------------------+
|    Metadata     |  元数据(可变长度)，包含服务名、方法名等信息
+------------------+
|    Payload      |  消息体(可变长度)，包含请求参数或响应结果
+------------------+
```

在Go代码中，消息结构定义如下：

```go
type Message struct {
    Header   Header    // 消息头部
    Metadata *Metadata // 元数据
    Payload  []byte    // 消息体
}
```

### 元数据 (Metadata)

元数据包含了RPC调用的核心信息，如服务名称、方法名称、错误信息等：

```go
type Metadata struct {
    ServiceName string            // 服务名称
    MethodName  string            // 方法名称
    Error       string            // 错误信息(仅响应消息使用)
    Extra       map[string]string // 额外的元数据，如trace_id等
}
```

元数据支持用户自定义扩展字段，可以通过`Extra`字段添加链路追踪ID、认证信息等附加数据。

## 协议头 (Header)

协议头是fyerrpc消息的固定部分，包含了处理消息所需的所有控制信息，采用固定长度的二进制格式。

### 头部结构

协议头总共22个字节，按字节划分的格式如下：

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

在Go中的定义：

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

### 头部字段详解

1. **魔数 (Magic Number)** - 2字节
    - 固定值：`0x3f3f`
    - 作用：快速校验是否为有效的fyerrpc消息，避免处理错误的消息

2. **版本 (Version)** - 1字节
    - 当前值：`0x01`
    - 作用：支持协议升级和向后兼容

3. **消息类型 (Message Type)** - 1字节
    - 请求消息：`0x01`
    - 响应消息：`0x02`
    - 作用：区分请求和响应消息

4. **压缩类型 (Compress Type)** - 1字节
    - 不压缩：`0x00`
    - Gzip压缩：`0x01`
    - 作用：指示消息体是否压缩及使用的压缩算法

5. **序列化类型 (Serialization Type)** - 1字节
    - JSON序列化：`0x01`
    - Protobuf序列化：`0x02`
    - 作用：指定元数据和消息体的序列化方式

6. **消息ID (Message ID)** - 8字节
    - 作用：唯一标识一个RPC请求，用于请求和响应的配对，支持异步调用和多路复用

7. **元数据长度 (Metadata Size)** - 4字节
    - 作用：指定元数据部分的字节长度

8. **消息体长度 (Payload Size)** - 4字节
    - 作用：指定消息体部分的字节长度

## 消息编解码

fyerrpc使用Protocol接口定义了消息的编码和解码行为：

```go
type Protocol interface {
    EncodeMessage(message *Message, writer io.Writer) error
    DecodeMessage(reader io.Reader) (*Message, error)
}
```

### 默认协议实现

`DefaultProtocol`是框架提供的标准实现，它按照二进制格式编码和解码消息：

```go
// DefaultProtocol 默认协议实现
type DefaultProtocol struct{}

// EncodeMessage 编码消息
func (p *DefaultProtocol) EncodeMessage(message *Message, writer io.Writer) error {
    // 写入头部各个字段
    if err := binary.Write(writer, binary.BigEndian, message.Header.MagicNumber); err != nil {
        return err
    }
    // ... 写入其他头部字段 ...

    // 序列化元数据
    var metadataBytes []byte
    var err error
    if message.Metadata != nil {
        codec := GetCodecByType(message.Header.SerializationType)
        if codec == nil {
            return ErrUnsupportedSerializer
        }

        metadataBytes, err = codec.Encode(message.Metadata)
        if err != nil {
            return err
        }
    }

    // 写入元数据长度
    message.Header.MetadataSize = uint32(len(metadataBytes))
    if err := binary.Write(writer, binary.BigEndian, message.Header.MetadataSize); err != nil {
        return err
    }

    // 写入消息体长度
    message.Header.PayloadSize = uint32(len(message.Payload))
    if err := binary.Write(writer, binary.BigEndian, message.Header.PayloadSize); err != nil {
        return err
    }

    // 写入元数据
    if len(metadataBytes) > 0 {
        if _, err := writer.Write(metadataBytes); err != nil {
            return err
        }
    }

    // 写入消息体
    if len(message.Payload) > 0 {
        if _, err := writer.Write(message.Payload); err != nil {
            return err
        }
    }

    return nil
}
```

解码过程是编码的逆过程：

```go
// DecodeMessage 解码消息
func (p *DefaultProtocol) DecodeMessage(reader io.Reader) (*Message, error) {
    message := &Message{
        Header: Header{},
    }

    // 读取头部各个字段
    if err := binary.Read(reader, binary.BigEndian, &message.Header.MagicNumber); err != nil {
        return nil, err
    }
    if message.Header.MagicNumber != MagicNumber {
        return nil, ErrInvalidMagic
    }

    // ... 读取其他头部字段 ...

    // 读取元数据
    if message.Header.MetadataSize > 0 {
        metadataBytes := make([]byte, message.Header.MetadataSize)
        if _, err := io.ReadFull(reader, metadataBytes); err != nil {
            return nil, err
        }

        codec := GetCodecByType(message.Header.SerializationType)
        if codec == nil {
            return nil, ErrUnsupportedSerializer
        }

        message.Metadata = &Metadata{}
        if err := codec.Decode(metadataBytes, message.Metadata); err != nil {
            return nil, err
        }
    }

    // 读取消息体
    if message.Header.PayloadSize > 0 {
        payload := make([]byte, message.Header.PayloadSize)
        if _, err := io.ReadFull(reader, payload); err != nil {
            return nil, err
        }
        message.Payload = payload
    }

    return message, nil
}
```

## 自定义协议

fyerrpc支持自定义协议，您可以扩展或完全替换默认的协议实现。

### 扩展默认协议

扩展默认协议最简单的方式是在现有协议基础上添加功能：

```go
// EnhancedProtocol 扩展默认协议，添加加密功能
type EnhancedProtocol struct {
    DefaultProtocol
    encryptionKey []byte
}

// EncodeMessage 重写编码方法，添加加密
func (p *EnhancedProtocol) EncodeMessage(message *Message, writer io.Writer) error {
    // 加密消息体
    if len(message.Payload) > 0 {
        encrypted, err := encrypt(message.Payload, p.encryptionKey)
        if err != nil {
            return err
        }
        message.Payload = encrypted
    }
    
    // 调用默认实现完成编码
    return p.DefaultProtocol.EncodeMessage(message, writer)
}

// DecodeMessage 重写解码方法，添加解密
func (p *EnhancedProtocol) DecodeMessage(reader io.Reader) (*Message, error) {
    // 先使用默认实现解码
    message, err := p.DefaultProtocol.DecodeMessage(reader)
    if err != nil {
        return nil, err
    }
    
    // 解密消息体
    if len(message.Payload) > 0 {
        decrypted, err := decrypt(message.Payload, p.encryptionKey)
        if err != nil {
            return nil, err
        }
        message.Payload = decrypted
    }
    
    return message, nil
}

// 创建加密协议实例
func NewEncryptedProtocol(key []byte) *EnhancedProtocol {
    return &EnhancedProtocol{
        encryptionKey: key,
    }
}
```

### 实现全新协议

如果默认协议不满足需求，可以自己实现完全不同的协议格式：

```go
// CompactProtocol 实现更紧凑的协议格式
type CompactProtocol struct{}

// EncodeMessage 使用紧凑格式编码消息
func (p *CompactProtocol) EncodeMessage(message *Message, writer io.Writer) error {
    // 实现紧凑编码逻辑
    // 例如：使用变长整数编码、位压缩等技术减少协议开销
    
    // 示例：使用varint编码消息ID
    var buf [10]byte
    n := binary.PutUvarint(buf[:], message.Header.MessageID)
    if _, err := writer.Write(buf[:n]); err != nil {
        return err
    }
    
    // ... 编码其他字段 ...
    
    return nil
}

// DecodeMessage 解码紧凑格式消息
func (p *CompactProtocol) DecodeMessage(reader io.Reader) (*Message, error) {
    message := &Message{
        Header: Header{},
    }
    
    // 实现紧凑解码逻辑
    // 示例：使用varint解码消息ID
    messageID, err := binary.ReadUvarint(reader.(io.ByteReader))
    if err != nil {
        return nil, err
    }
    message.Header.MessageID = messageID
    
    // ... 解码其他字段 ...
    
    return message, nil
}
```

## 协议使用示例

### 基本使用

```go
// 创建协议实例
protocol := &protocol.DefaultProtocol{}

// 创建请求消息
message := &protocol.Message{
    Header: protocol.Header{
        MagicNumber:       protocol.MagicNumber,
        Version:           1,
        MessageType:       protocol.TypeRequest,
        CompressType:      protocol.CompressTypeNone,
        SerializationType: protocol.SerializationTypeJSON,
        MessageID:         1234567890,
    },
    Metadata: &protocol.Metadata{
        ServiceName: "UserService",
        MethodName:  "GetUser",
        Extra: map[string]string{
            "trace_id": "abc123",
            "user_id":  "1001",
        },
    },
    Payload: []byte(`{"id": 1}`),
}

// 编码消息
buf := new(bytes.Buffer)
err := protocol.EncodeMessage(message, buf)
if err != nil {
    log.Fatalf("编码消息失败: %v", err)
}

// 解码消息
decoded, err := protocol.DecodeMessage(buf)
if err != nil {
    log.Fatalf("解码消息失败: %v", err)
}

fmt.Printf("服务名: %s, 方法名: %s\n", 
    decoded.Metadata.ServiceName, 
    decoded.Metadata.MethodName)
```

### 使用自定义协议

```go
// 创建加密协议实例
encryptedProtocol := NewEncryptedProtocol([]byte("secret-key-12345"))

// 编码和解码消息
buf := new(bytes.Buffer)
err := encryptedProtocol.EncodeMessage(message, buf)
if err != nil {
    log.Fatalf("加密编码消息失败: %v", err)
}

decoded, err := encryptedProtocol.DecodeMessage(buf)
if err != nil {
    log.Fatalf("解密解码消息失败: %v", err)
}
```