package codec

import (
	"encoding/json"
)

func init() {
	RegisterCodec(JSON, &JsonCodec{})
}

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
