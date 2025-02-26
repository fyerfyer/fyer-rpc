package rpc

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/fyerfyer/fyer-rpc/protocol"
	"github.com/fyerfyer/fyer-rpc/protocol/codec"
)

type Client struct {
	conn      *Connection
	messageID uint64
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, NewRPCError(ErrCodeInternal, "failed to connect: "+err.Error())
	}

	return &Client{
		conn: NewConnection(conn),
	}, nil
}

func (c *Client) Call(serviceName, methodName string, args interface{}) ([]byte, error) {
	// 序列化请求参数
	serializer := codec.GetCodec(codec.JSON) // 默认使用JSON
	argBytes, err := serializer.Encode(args)
	if err != nil {
		return nil, NewRPCError(ErrCodeInvalidParam, "failed to marshal request: "+err.Error())
	}

	// 构造元数据
	metadata := &protocol.Metadata{
		ServiceName: serviceName,
		MethodName:  methodName,
	}

	// 生成消息ID
	messageID := atomic.AddUint64(&c.messageID, 1)

	// 发送请求
	err = c.conn.Write(
		serviceName,
		methodName,
		protocol.TypeRequest,
		protocol.SerializationTypeJSON,
		messageID,
		metadata,
		argBytes,
	)
	if err != nil {
		return nil, NewRPCError(ErrCodeInternal, "failed to send request: "+err.Error())
	}

	// 接收响应
	resp, err := c.conn.Read()
	if err != nil {
		return nil, NewRPCError(ErrCodeInternal, "failed to receive response: "+err.Error())
	}

	// 检查响应中的错误
	if resp.Metadata != nil && resp.Metadata.Error != "" {
		return nil, NewRPCError(ErrCodeInternal, resp.Metadata.Error)
	}

	return resp.Payload, nil
}

// CallWithTimeout 带超时的RPC调用
func (c *Client) CallWithTimeout(ctx context.Context, serviceName, methodName string, args interface{}) ([]byte, error) {
	done := make(chan struct{})
	var result []byte
	var err error

	go func() {
		result, err = c.Call(serviceName, methodName, args)
		close(done)
	}()

	select {
	case <-ctx.Done():
		// 不再关闭连接，而是让连接池处理
		return nil, NewRPCError(ErrCodeInternal, "request timeout")
	case <-done:
		return result, err
	}
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return NewRPCError(ErrCodeInternal, "failed to close connection: "+err.Error())
	}
	return nil
}
