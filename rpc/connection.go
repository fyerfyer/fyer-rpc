package rpc

import (
	"net"

	"github.com/fyerfyer/fyer-rpc/protocol"
)

// Connection 包装底层连接，提供协议层面的读写能力
type Connection struct {
	conn     net.Conn
	protocol protocol.Protocol
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:     conn,
		protocol: &protocol.DefaultProtocol{},
	}
}

func (c *Connection) Write(serviceName, methodName string, messageType uint8, serializationType uint8, messageID uint64, metadata *protocol.Metadata, payload []byte) error {
	message := &protocol.Message{
		Header: protocol.Header{
			MagicNumber:       protocol.MagicNumber,
			Version:           1,
			MessageType:       messageType,
			CompressType:      protocol.CompressTypeNone, // 暂时不启用压缩
			SerializationType: serializationType,
			MessageID:         messageID,
		},
		Metadata: metadata,
		Payload:  payload,
	}

	return c.protocol.EncodeMessage(message, c.conn)
}

func (c *Connection) Read() (*protocol.Message, error) {
	return c.protocol.DecodeMessage(c.conn)
}

func (c *Connection) Close() error {
	return c.conn.Close()
}
