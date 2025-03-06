package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fyerfyer/fyer-rpc/api"
	"github.com/fyerfyer/fyer-rpc/protocol"
	"github.com/fyerfyer/fyer-rpc/utils"
)

func main() {
	// 配置日志
	utils.SetDefaultLogger(utils.NewLogger(utils.InfoLevel, os.Stdout))

	// 创建服务器配置
	options := &api.ServerOptions{
		Address:       ":8000",                        // 服务监听地址
		SerializeType: protocol.SerializationTypeJSON, // 使用JSON序列化
	}

	// 创建服务器
	server := api.NewServer(options)

	// 注册服务
	greeter := &GreeterService{}
	err := server.Register(greeter)
	if err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 启动服务器
	log.Println("Starting RPC server on", options.Address)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 等待终止信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down server...")
	server.Stop()
}
