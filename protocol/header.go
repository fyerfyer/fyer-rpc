package protocol

// MagicNumber 协议头部魔数，用于快速校验协议
const MagicNumber uint16 = 0x3f3f

// 消息类型
const (
	TypeRequest  = uint8(0x01) // 请求消息
	TypeResponse = uint8(0x02) // 响应消息
)

// 序列化类型
const (
	SerializationTypeJSON     = uint8(0x01) // JSON序列化
	SerializationTypeProtobuf = uint8(0x02) // Protobuf序列化
)

// 压缩类型
const (
	CompressTypeNone = uint8(0x00) // 不压缩
	CompressTypeGzip = uint8(0x01) // Gzip压缩
)

// Header 协议头部结构
// 协议头部格式（按字节划分）:
// +-----------------------------------------------+
// |  magic number   |  version    |  msg type     |
// +-----------------------------------------------+
// |  2 bytes        |  1 byte     |  1 byte       |
// +-----------------------------------------------+
// |  compress type  |  serial type|  message id   |
// +-----------------------------------------------+
// |  1 byte         |  1 byte     |  8 bytes      |
// +-----------------------------------------------+
// |  metadata size  |  payload size               |
// +-----------------------------------------------+
// |  4 bytes        |  4 bytes                    |
// +-----------------------------------------------+
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

// HeaderSize 头部固定长度：22字节
const HeaderSize = 22
