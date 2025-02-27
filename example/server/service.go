package main

import (
	"context"
	"github.com/fyerfyer/fyer-rpc/naming"
	"log"
	"time"

	"github.com/fyerfyer/fyer-rpc/example/helloworld"
	"github.com/fyerfyer/fyer-rpc/registry"
	_ "github.com/fyerfyer/fyer-rpc/registry/etcd"
)

// GreetServer 是 GreetService 的服务端实现
type GreetServer struct {
	*helloworld.GreetServiceImpl
	registry registry.Registry
	instance *naming.Instance
}

// NewGreetServer 创建新的 GreetServer 实例
func NewGreetServer(reg registry.Registry, addr string) *GreetServer {
	return &GreetServer{
		GreetServiceImpl: helloworld.NewGreetService(),
		registry:         reg,
		instance: &naming.Instance{
			ID:      addr, // 使用地址作为实例ID
			Service: "GreetService",
			Version: "1.0.0",
			Address: addr,
			Status:  naming.StatusEnabled,
			Metadata: map[string]string{
				"protocol": "tcp",
				"weight":   "100",
			},
		},
	}
}

// Start 启动服务并注册到注册中心
func (s *GreetServer) Start(ctx context.Context) error {
	// 注册服务实例
	if err := s.registry.Register(ctx, s.instance); err != nil {
		return err
	}

	// 启动心跳
	go s.heartbeat(ctx)

	return nil
}

// Stop 停止服务并从注册中心注销
func (s *GreetServer) Stop(ctx context.Context) error {
	return s.registry.Deregister(ctx, s.instance)
}

// heartbeat 定期发送心跳
func (s *GreetServer) heartbeat(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 5):
			if err := s.registry.Heartbeat(ctx, s.instance); err != nil {
				log.Printf("Failed to send heartbeat: %v", err)
			}
		}
	}
}
