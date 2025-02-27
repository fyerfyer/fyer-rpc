package helloworld

import (
	"context"
	"sync"
	"time"
)

// GreetServiceImpl 问候服务实现
type GreetServiceImpl struct {
	mu sync.RWMutex
	// 统计信息
	totalGreets   int64
	greetsByName  map[string]int64
	lastGreetTime int64
}

// NewGreetService 创建问候服务实例
func NewGreetService() *GreetServiceImpl {
	return &GreetServiceImpl{
		greetsByName: make(map[string]int64),
	}
}

// SayHello 实现问候方法
func (s *GreetServiceImpl) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 更新统计信息
	s.totalGreets++
	s.greetsByName[req.Name]++
	now := time.Now().UnixNano()
	s.lastGreetTime = now

	// 构造问候语
	greeting := req.Greeting
	if greeting == "" {
		greeting = "Hello"
	}
	message := greeting + ", " + req.Name + "!"

	return &HelloResponse{
		Message:   message,
		GreetTime: now,
	}, nil
}

// GetGreetStats 实现统计方法
func (s *GreetServiceImpl) GetGreetStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	response := &StatsResponse{
		TotalGreets:   s.totalGreets,
		GreetsByName:  make(map[string]int64),
		LastGreetTime: s.lastGreetTime,
	}

	// 如果指定了名字，只返回该用户的统计
	if req.Name != "" {
		if count, ok := s.greetsByName[req.Name]; ok {
			response.GreetsByName[req.Name] = count
		}
	} else {
		// 否则返回所有用户的统计
		for name, count := range s.greetsByName {
			response.GreetsByName[name] = count
		}
	}

	return response, nil
}
