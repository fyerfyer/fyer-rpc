package api

import (
	"context"
	"errors"
	"time"

	"github.com/fyerfyer/fyer-rpc/protocol"
	"github.com/fyerfyer/fyer-rpc/rpc"
)

// 错误定义
var (
	ErrClientClosed    = errors.New("client is closed")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNoAddress       = errors.New("no server address specified")
)

// Client 定义了RPC客户端接口
type Client interface {
	// Call 同步调用远程服务
	Call(ctx context.Context, service, method string, req interface{}, resp interface{}) error

	// Close 关闭客户端连接
	Close() error
}

// ClientOptions 客户端配置选项
type ClientOptions struct {
	Address       string        // 服务器地址
	Timeout       time.Duration // 请求超时
	PoolSize      int           // 连接池大小
	SerializeType uint8         // 序列化类型
}

// 默认客户端配置
var DefaultClientOptions = &ClientOptions{
	Timeout:       time.Second * 5,
	PoolSize:      10,
	SerializeType: protocol.SerializationTypeJSON,
}

// 简单客户端实现
type simpleClient struct {
	pool    *rpc.ConnPool
	options *ClientOptions
	closed  bool
}

// NewClient 创建新的RPC客户端
func NewClient(options *ClientOptions) (Client, error) {
	if options == nil {
		options = DefaultClientOptions
	}

	if options.Address == "" {
		return nil, ErrNoAddress
	}

	pool := rpc.NewConnPool(options.Address, options.PoolSize, time.Minute*5)

	return &simpleClient{
		pool:    pool,
		options: options,
	}, nil
}

// Call 实现同步调用
func (c *simpleClient) Call(ctx context.Context, service, method string, req interface{}, resp interface{}) error {
	if c.closed {
		return ErrClientClosed
	}

	if req == nil || resp == nil {
		return ErrInvalidArgument
	}

	// 添加超时
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.options.Timeout)
		defer cancel()
	}

	// 从连接池获取连接
	client, err := c.pool.Get()
	if err != nil {
		return err
	}
	defer c.pool.Put(client)

	// 执行调用
	data, err := client.Call(service, method, req)
	if err != nil {
		return err
	}

	// 反序列化响应
	decoder := protocol.GetCodecByType(c.options.SerializeType)
	if decoder == nil {
		return errors.New("no available decoder")
	}

	return decoder.Decode(data, resp)
}

// Close 关闭客户端
func (c *simpleClient) Close() error {
	if c.closed {
		return nil
	}

	c.closed = true
	if c.pool != nil {
		c.pool.Close()
	}

	return nil
}
