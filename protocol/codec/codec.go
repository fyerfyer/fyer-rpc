package codec

import "errors"

var (
	ErrInvalidMessage = errors.New("message type is invalid")
)

// Codec 定义了序列化和反序列化的接口
type Codec interface {
	// Encode 将对象序列化为字节数组
	Encode(v interface{}) ([]byte, error)

	// Decode 将字节数组反序列化为对象
	Decode(data []byte, v interface{}) error

	// Name 返回编解码器的名称
	Name() string
}

// Type 定义了支持的序列化类型
type Type uint8

const (
	JSON Type = iota
	Protobuf
)

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
