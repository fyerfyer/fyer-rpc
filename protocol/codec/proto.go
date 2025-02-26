package codec

import (
	"google.golang.org/protobuf/proto"
)

func init() {
	RegisterCodec(Protobuf, &ProtobufCodec{})
}

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
