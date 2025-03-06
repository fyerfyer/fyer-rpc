package api

import (
	"net"

	"github.com/fyerfyer/fyer-rpc/protocol"
	"github.com/fyerfyer/fyer-rpc/registry/etcd"
	"github.com/fyerfyer/fyer-rpc/rpc"
	"github.com/fyerfyer/fyer-rpc/utils"
)

// Server 定义了RPC服务器接口
type Server interface {
	// Register 注册服务
	Register(service interface{}) error

	// Start 启动服务器
	Start() error

	// Stop 停止服务器
	Stop() error

	// Address 获取服务器监听地址
	Address() string
}

// ServerOptions 服务器配置选项
type ServerOptions struct {
	Address        string   // 服务监听地址，默认":8000"
	SerializeType  uint8    // 序列化类型，默认JSON
	EnableRegistry bool     // 是否启用服务注册
	RegistryAddrs  []string // 注册中心地址
	ServiceName    string   // 服务名称
	ServiceVersion string   // 服务版本
}

// 默认服务器配置
var DefaultServerOptions = &ServerOptions{
	Address:        ":8000",
	SerializeType:  protocol.SerializationTypeJSON,
	EnableRegistry: false,
	ServiceVersion: "1.0.0",
}

// 简单服务器实现
type simpleServer struct {
	server   *rpc.Server
	options  *ServerOptions
	registry *etcd.EtcdRegistry
	listener net.Listener
	started  bool
	stopped  bool
}

// NewServer 创建新的RPC服务器
func NewServer(options *ServerOptions) Server {
	if options == nil {
		options = DefaultServerOptions
	}

	server := rpc.NewServer()
	server.SetSerializationType(options.SerializeType)

	return &simpleServer{
		server:  server,
		options: options,
	}
}

// Register 注册服务
func (s *simpleServer) Register(service interface{}) error {
	return s.server.RegisterService(service)
}

// Start 启动服务器
func (s *simpleServer) Start() error {
	if s.started {
		return nil
	}

	// 初始化注册中心
	if s.options.EnableRegistry && len(s.options.RegistryAddrs) > 0 {
		reg, err := etcd.New(
			etcd.WithEndpoints(s.options.RegistryAddrs),
			etcd.WithTTL(30),
		)
		if err != nil {
			utils.Warn("Failed to init registry: %v", err)
		} else {
			s.registry = reg
		}
	}

	// 不创建监听器，直接让底层Server去创建
	s.started = true
	go func() {
		if err := s.server.Start(s.options.Address); err != nil {
			utils.Error("Server error: %v", err)
		}
	}()

	utils.Info("Server started at %s", s.options.Address)
	return nil
}

// Stop 停止服务器
func (s *simpleServer) Stop() error {
	if !s.started || s.stopped {
		return nil
	}

	s.stopped = true

	// 关闭注册中心
	if s.registry != nil {
		s.registry.Close()
	}

	// 通知底层服务器停止
	if s.server != nil {
		// 如果底层Server有停止方法，调用它
		if stopServer, ok := interface{}(s.server).(interface{ Stop() error }); ok {
			return stopServer.Stop()
		}
	}

	return nil
}

// Address 获取服务器监听地址
func (s *simpleServer) Address() string {
	return s.options.Address
}
