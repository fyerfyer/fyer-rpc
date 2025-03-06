package main

import (
	"context"
	"fmt"
)

// 请求结构体
type HelloRequest struct {
	Name string `json:"name"`
}

// 响应结构体
type HelloResponse struct {
	Message string `json:"message"`
}

// 直接定义服务结构体
type GreeterService struct{}

// SayHello 方法实现
func (s *GreeterService) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.Name),
	}, nil
}

func (s *GreeterService) SayGoodbye(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{
		Message: fmt.Sprintf("Goodbye, %s!", req.Name),
	}, nil
}
