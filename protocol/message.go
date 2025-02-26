package protocol

// Message 完整的消息结构
// 消息格式:
// +------------------+
// |     Header      |  消息头部（固定22字节）
// +------------------+
// |    Metadata     |  元数据(可变长度)，包含服务名、方法名等信息
// +------------------+
// |    Payload      |  消息体(可变长度)，包含请求参数或响应结果
// +------------------+
type Message struct {
	Header   Header    // 消息头部
	Metadata *Metadata // 元数据
	Payload  []byte    // 消息体
}

// Error 实现接口
func (m Message) Error() string {
	return m.Metadata.Error
}

// Metadata 元数据结构
// 包含RPC调用的必要信息和可选的链路追踪信息
type Metadata struct {
	ServiceName string            // 服务名称
	MethodName  string            // 方法名称
	Error       string            // 错误信息(仅响应消息使用)
	Extra       map[string]string // 额外的元数据，如trace_id等
}

// RequestMessage 请求消息
// type RequestMessage struct {
// 	*Message
// }

// // ResponseMessage 响应消息
// type ResponseMessage struct {
// 	*Message
// }
