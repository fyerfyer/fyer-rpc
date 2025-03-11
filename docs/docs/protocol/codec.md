# codec模块

`codec.go`文件定义了 fyer-rpc 框架中的编解码器（Codec）接口和管理系统，负责处理消息的序列化和反序列化过程。

## `Codec`接口

该模块定义了编解码器的核心接口：

```go
// Codec 定义了序列化和反序列化的接口
type Codec interface {
    // Encode 将对象序列化为字节数组
    Encode(v interface{}) ([]byte, error)

    // Decode 将字节数组反序列化为对象
    Decode(data []byte, v interface{}) error

    // Name 返回编解码器的名称
    Name() string
}
```

## 编解码器注册管理

```go
var (
    codecs = make(map[Type]Codec)
)

// RegisterCodec 注册编解码器
func RegisterCodec(t Type, codec Codec) {
    codecs[t] = codec
}

// GetCodec 获取编解码器
func GetCodec(t Type) Codec {
    return codecs[t]
}
```

## 使用示例

```go
// 1. 获取 JSON 编解码器
jsonCodec := codec.GetCodec(codec.JSON)
if jsonCodec == nil {
    log.Fatal("failed to register JSON codec")
}

// 2. 定义一个数据结构
type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// 3. 序列化对象
person := Person{Name: "fyerfyer", Age: 18}
data, err := jsonCodec.Encode(person)
if err != nil {
    log.Fatalf("failed to encode: %v", err)
}
fmt.Printf("encode result: %s\n", data)

// 4. 反序列化对象
var decodedPerson Person
err = jsonCodec.Decode(data, &decodedPerson)
if err != nil {
    log.Fatalf("failed to decode: %v", err)
}
fmt.Printf("decode result: %+v\n", decodedPerson)

// 5. 获取编解码器名称
fmt.Printf("codec name: %s\n", jsonCodec.Name())
```