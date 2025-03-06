package main

import (
	"context"
	"log"
	"time"

	"github.com/fyerfyer/fyer-rpc/api"
	"github.com/fyerfyer/fyer-rpc/protocol"
)

func main() {
	// 创建客户端配置
	options := &api.ClientOptions{
		Address:       "localhost:8000",               // 服务器地址
		Timeout:       time.Second * 5,                // 请求超时
		SerializeType: protocol.SerializationTypeJSON, // 使用JSON序列化
	}

	// 创建客户端
	client, err := api.NewClient(options)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 创建请求对象
	request := &HelloRequest{
		Name: "World",
	}

	// 创建响应对象
	response := &HelloResponse{}

	// 调用远程服务
	// 注意：服务名称就是结构体的名称 "GreeterService"
	err = client.Call(context.Background(), "GreeterService", "SayHello", request, response)
	if err != nil {
		log.Fatalf("RPC call failed: %v", err)
	}

	// 打印响应
	log.Printf("SayHello response: %s", response.Message)

	// 调用另一个方法
	err = client.Call(context.Background(), "GreeterService", "SayGoodbye", request, response)
	if err != nil {
		log.Fatalf("RPC call failed: %v", err)
	}

	// 打印响应
	log.Printf("SayGoodbye response: %s", response.Message)
}
