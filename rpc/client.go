package rpc

import (
	"context"
	"encoding/json"
	"net"
)

type Client struct {
	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, NewRPCError(ErrCodeInternal, "failed to connect: "+err.Error())
	}

	return &Client{
		conn:    conn,
		encoder: json.NewEncoder(conn),
		decoder: json.NewDecoder(conn),
	}, nil
}

func (c *Client) Call(serviceName, methodName string, args interface{}) ([]byte, error) {
	// 序列化参数
	argBytes, err := json.Marshal(args)
	if err != nil {
		return nil, NewRPCError(ErrCodeInvalidParam, "failed to marshal request: "+err.Error())
	}

	// 发送请求
	req := Request{
		ServiceName: serviceName,
		MethodName:  methodName,
		Args:        argBytes,
	}

	if err := c.encoder.Encode(req); err != nil {
		return nil, NewRPCError(ErrCodeInternal, "failed to send request: "+err.Error())
	}

	// 接收响应
	var resp Response
	if err := c.decoder.Decode(&resp); err != nil {
		return nil, NewRPCError(ErrCodeInternal, "failed to receive response: "+err.Error())
	}

	// 检查响应中的错误
	if resp.Error != "" {
		return nil, NewRPCError(ErrCodeInternal, resp.Error)
	}

	if resp.Data == nil {
		return nil, NewRPCError(ErrCodeInternal, "empty response data")
	}

	return resp.Data, nil
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
