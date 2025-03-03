package rpc

import (
	"context"
	"github.com/fyerfyer/fyer-rpc/cluster/failover"
	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/fyerfyer/fyer-rpc/protocol"
	"github.com/fyerfyer/fyer-rpc/protocol/codec"
	"net"
	"sync/atomic"
)

type Client struct {
	conn            *Connection
	messageID       uint64
	failoverConfig  *failover.Config                 // 故障转移配置
	failoverHandler *failover.DefaultFailoverHandler // 故障转移处理器
	enableFailover  bool                             // 是否启用故障转移
}

// ClientOption 客户端配置选项
type ClientOption func(*Client)

// WithFailover 启用故障转移功能
func WithFailover(config *failover.Config) ClientOption {
	return func(c *Client) {
		c.failoverConfig = config
		c.enableFailover = true
		handler, err := failover.NewFailoverHandler(config)
		if err == nil {
			c.failoverHandler = handler
		}
	}
}

func NewClient(address string, options ...ClientOption) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, NewRPCError(ErrCodeInternal, "failed to connect: "+err.Error())
	}

	client := &Client{
		conn:           NewConnection(conn),
		enableFailover: false,
	}

	// 应用配置选项
	for _, option := range options {
		option(client)
	}

	return client, nil
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

// CallWithFailover 带故障转移功能的RPC调用
func (c *Client) CallWithFailover(ctx context.Context, serviceName, methodName string, args interface{}, instances []*naming.Instance) ([]byte, error) {
	if !c.enableFailover || c.failoverHandler == nil || len(instances) == 0 {
		// 未启用故障转移或没有可用实例，直接调用原始方法
		return c.CallWithTimeout(ctx, serviceName, methodName, args)
	}

	// 封装RPC调用操作
	operation := func(ctx context.Context, instance *naming.Instance) error {
		// 创建到具体实例的新连接
		client, err := NewClient(instance.Address)
		if err != nil {
			return err
		}
		defer client.Close()

		// 执行RPC调用
		_, err = client.CallWithTimeout(ctx, serviceName, methodName, args)
		return err
	}

	// 执行带故障转移的调用
	result, err := c.failoverHandler.Execute(ctx, instances, operation)
	if err != nil {
		return nil, NewRPCError(ErrCodeInternal, "failover failed: "+err.Error())
	}

	// 如果故障转移成功，通过成功的实例再执行一次调用返回结果
	if result.Success {
		client, err := NewClient(result.Instance.Address)
		if err != nil {
			return nil, NewRPCError(ErrCodeInternal, "failed to connect to selected instance: "+err.Error())
		}
		defer client.Close()

		return client.CallWithTimeout(ctx, serviceName, methodName, args)
	}

	return nil, NewRPCError(ErrCodeInternal, "no available instances after failover attempts")
}

// IsFailoverEnabled 检查是否启用了故障转移功能
func (c *Client) IsFailoverEnabled() bool {
	return c.enableFailover && c.failoverHandler != nil
}

// GetFailoverHandler 获取故障转移处理器
func (c *Client) GetFailoverHandler() *failover.DefaultFailoverHandler {
	return c.failoverHandler
}

func (c *Client) Close() error {
	return c.conn.Close()
}
