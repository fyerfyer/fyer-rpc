# Serilization

序列化是RPC框架中的关键环节，负责将数据结构转换为可传输的字节流，以及将接收到的字节流还原为数据结构。

## 序列化接口

fyerrpc通过`codec.Codec`接口定义了序列化和反序列化的标准操作：

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

所有的序列化实现都必须遵循这个接口，确保框架可以统一处理不同的序列化方式。

## 支持的序列化类型

fyerrpc目前支持两种序列化类型，分别是：

```go
// Type 定义了支持的序列化类型
type Type uint8

const (
    JSON Type = iota      // JSON序列化
    Protobuf              // Protobuf序列化
)
```

在协议层面，这些类型对应的值为：

```go
// 序列化类型
const (
    SerializationTypeJSON     = uint8(0x01) // JSON序列化
    SerializationTypeProtobuf = uint8(0x02) // Protobuf序列化
)
```

## JSON序列化

### 实现方式

fyerrpc的JSON序列化是通过Go标准库的`encoding/json`包实现的：

```go
// JsonCodec 实现了 Codec 接口
type JsonCodec struct{}

// Encode 将对象序列化为 JSON 字节数组
func (c *JsonCodec) Encode(v interface{}) ([]byte, error) {
    return json.Marshal(v)
}

// Decode 将 JSON 字节数组反序列化为对象
func (c *JsonCodec) Decode(data []byte, v interface{}) error {
    return json.Unmarshal(data, v)
}

// Name 返回编解码器的名称
func (c *JsonCodec) Name() string {
    return "json"
}
```

### 使用方法

在客户端或服务端配置中指定使用JSON序列化：

```go
// 服务端
server := api.NewServer(&api.ServerOptions{
    Address:       ":8000",
    SerializeType: protocol.SerializationTypeJSON, // 使用JSON序列化
})

// 客户端
client, err := api.NewClient(&api.ClientOptions{
    Address:       "localhost:8000",
    SerializeType: protocol.SerializationTypeJSON, // 使用JSON序列化
})
```

如果使用底层API，可以直接获取JSON编解码器：

```go
import "github.com/fyerfyer/fyer-rpc/protocol/codec"

// 获取JSON编解码器
jsonCodec := codec.GetCodec(codec.JSON)

// 序列化
data, err := jsonCodec.Encode(myStruct)
if err != nil {
    log.Fatalf("Failed to encode: %v", err)
}

// 反序列化
var result MyStruct
err = jsonCodec.Decode(data, &result)
if err != nil {
    log.Fatalf("Failed to decode: %v", err)
}
```

## Protobuf序列化

### 实现方式

fyerrpc的Protobuf序列化是通过`google.golang.org/protobuf/proto`包实现的：

```go
// ProtobufCodec 实现了 Codec 接口
type ProtobufCodec struct{}

// Encode 将对象序列化为 Protobuf 字节数组
func (c *ProtobufCodec) Encode(v interface{}) ([]byte, error) {
    // 类型断言确保v是proto.Message类型
    if pm, ok := v.(proto.Message); ok {
        return proto.Marshal(pm)
    }
    return nil, ErrInvalidMessage
}

// Decode 将 Protobuf 字节数组反序列化为对象
func (c *ProtobufCodec) Decode(data []byte, v interface{}) error {
    // 类型断言确保v是proto.Message类型
    if pm, ok := v.(proto.Message); ok {
        return proto.Unmarshal(data, pm)
    }
    return ErrInvalidMessage
}

// Name 返回编解码器的名称
func (c *ProtobufCodec) Name() string {
    return "protobuf"
}
```

### 使用方法

#### 定义Protocol Buffers

首先，需要创建`.proto`文件定义消息结构：

```protobuf
syntax = "proto3";
package example;

option go_package = "github.com/fyerfyer/fyer-rpc/example/proto";

message HelloRequest {
    string name = 1;
}

message HelloResponse {
    string message = 1;
}
```

然后使用`protoc`工具生成Go代码：

```bash
protoc --go_out=. --go_opt=paths=source_relative hello.proto
```

#### 在fyerrpc中使用Protobuf

在客户端或服务端配置中指定使用Protobuf序列化：

```go
// 服务端
server := api.NewServer(&api.ServerOptions{
    Address:       ":8000",
    SerializeType: protocol.SerializationTypeProtobuf, // 使用Protobuf序列化
})

// 客户端
client, err := api.NewClient(&api.ClientOptions{
    Address:       "localhost:8000",
    SerializeType: protocol.SerializationTypeProtobuf, // 使用Protobuf序列化
})
```

如果使用底层API，可以直接获取Protobuf编解码器：

```go
import (
    "github.com/fyerfyer/fyer-rpc/protocol/codec"
    "github.com/fyerfyer/fyer-rpc/example/proto"
)

// 获取Protobuf编解码器
pbCodec := codec.GetCodec(codec.Protobuf)

// 创建Protobuf消息
request := &proto.HelloRequest{
    Name: "World",
}

// 序列化
data, err := pbCodec.Encode(request)
if err != nil {
    log.Fatalf("Failed to encode: %v", err)
}

// 反序列化
response := &proto.HelloResponse{}
err = pbCodec.Decode(data, response)
if err != nil {
    log.Fatalf("Failed to decode: %v", err)
}
```

## 协议中的序列化

fyerrpc协议头中包含了序列化类型字段，用于标识消息体使用的序列化方式：

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

当接收到消息时，框架会根据头部的`SerializationType`字段选择合适的解码器：

```go
func GetCodecByType(serializationType uint8) codec.Codec {
    switch serializationType {
    case SerializationTypeJSON:
        return codec.GetCodec(codec.JSON)
    case SerializationTypeProtobuf:
        return codec.GetCodec(codec.Protobuf)
    default:
        return nil
    }
}
```

## 元数据序列化

fyerrpc中的元数据（Metadata）也需要序列化和反序列化：

```go
// Metadata 元数据结构
type Metadata struct {
    ServiceName string            // 服务名称
    MethodName  string            // 方法名称
    Error       string            // 错误信息(仅响应消息使用)
    Extra       map[string]string // 额外的元数据，如trace_id等
}
```

元数据的序列化方式由消息头的`SerializationType`字段指定，与消息体使用相同的序列化方式。

## 示例代码

### JSON序列化示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fyerfyer/fyer-rpc/api"
    "github.com/fyerfyer/fyer-rpc/protocol"
)

// 请求和响应结构体
type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

func main() {
    // 创建客户端，使用JSON序列化
    client, err := api.NewClient(&api.ClientOptions{
        Address:       "localhost:8000",
        SerializeType: protocol.SerializationTypeJSON,
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

    // 创建请求
    request := &HelloRequest{Name: "World"}
    response := &HelloResponse{}

    // 调用远程服务
    err = client.Call(context.Background(), "GreeterService", "SayHello", request, response)
    if err != nil {
        log.Fatalf("RPC call failed: %v", err)
    }

    fmt.Printf("Response: %s\n", response.Message)
}
```

### Protobuf序列化示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fyerfyer/fyer-rpc/api"
    "github.com/fyerfyer/fyer-rpc/protocol"
    "github.com/fyerfyer/fyer-rpc/example/proto"
)

func main() {
    // 创建客户端，使用Protobuf序列化
    client, err := api.NewClient(&api.ClientOptions{
        Address:       "localhost:8000",
        SerializeType: protocol.SerializationTypeProtobuf,
    })
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()

    // 创建Protobuf请求
    request := &proto.HelloRequest{Name: "World"}
    response := &proto.HelloResponse{}

    // 调用远程服务
    err = client.Call(context.Background(), "GreeterService", "SayHello", request, response)
    if err != nil {
        log.Fatalf("RPC call failed: %v", err)
    }

    fmt.Printf("Response: %s\n", response.Message)
}
```