package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/example/common"
)

// DefaultServerConfigs 服务器配置信息，用于客户端展示
var DefaultServerConfigs = []*common.ServerConfig{
	{
		ID:           "server-A",
		Address:      "localhost",
		Port:         8001,
		FailAfter:    100, // 处理100个请求后故障
		FailDuration: 10 * time.Second,
	},
	{
		ID:       "server-B",
		Address:  "localhost",
		Port:     8002,
		FailRate: 0.1, // 10%概率随机故障
	},
	{
		ID:      "server-C",
		Address: "localhost",
		Port:    8003,
	},
}

func main() {
	// 创建带故障转移功能的配置
	serverAddresses := []string{
		"localhost:8001",
		"localhost:8002",
		"localhost:8003",
	}
	clientConfig := CreateDefaultClientConfig(serverAddresses)

	// 创建带故障转移功能的客户端
	client, err := NewGreetClient("GreetService", clientConfig)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 创建上下文
	ctx := context.Background()

	// 启动健康监测器
	go monitorServerHealth(ctx, clientConfig, client.failoverManager)

	fmt.Println("\n=== Failover Demonstration Begins ===")
	fmt.Println("* Multiple server instances will simulate different failure scenarios:")
	fmt.Printf("  - Server A (Port 8001): Fails after processing %d requests for %v\n",
		DefaultServerConfigs[0].FailAfter,
		DefaultServerConfigs[0].FailDuration)
	fmt.Printf("  - Server B (Port 8002): %.0f%% chance of random failure\n",
		DefaultServerConfigs[1].FailRate*100)
	fmt.Printf("  - Server C (Port 8003): Runs normally\n")
	fmt.Println("* The client will automatically failover to healthy servers")
	fmt.Println("* Observe the failover events and request flow in the output")
	fmt.Println("==========================================")

	// 测试一般调用
	testBasicCalls(ctx, client)

	// 测试故障转移
	testFailover(ctx, client)

	// 测试熔断器
	testCircuitBreaker(ctx, client)

	// 测试并发调用
	testConcurrentCalls(ctx, client)

	// 显示最终指标
	time.Sleep(1 * time.Second)
	showFailoverMetrics(client)

	fmt.Println("\nDemonstration completed")
}

// 健康监测器
func monitorServerHealth(ctx context.Context, config *common.ClientConfig, failoverManager *FailoverManager) {
	if failoverManager == nil {
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			results := failoverManager.CheckAllServers(ctx)
			var healthy, unhealthy int
			for _, isHealthy := range results {
				if isHealthy {
					healthy++
				} else {
					unhealthy++
				}
			}
			log.Printf("Server health status: %d healthy, %d unhealthy", healthy, unhealthy)
		}
	}
}

// 测试基本调用
func testBasicCalls(ctx context.Context, client *GreetClient) {
	fmt.Println("\n=== Basic Call Test ===")

	for i := 0; i < 3; i++ {
		resp, err := client.SayHello(ctx, fmt.Sprintf("User%d", i), "Hello")
		if err != nil {
			log.Printf("Error: %v", err)
		} else {
			log.Printf("Response: %s", resp.Message)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// 测试故障转移
func testFailover(ctx context.Context, client *GreetClient) {
	fmt.Println("\n=== Failover Test ===")
	fmt.Println("* Sending multiple requests quickly to trigger failure on Server A")

	// 快速发送多个请求，触发服务器故障
	for i := 0; i < 120; i++ {
		resp, err := client.SayHello(ctx, fmt.Sprintf("User%d", i), "Hello")
		if err != nil {
			log.Printf("Error: %v", err)
		} else {
			log.Printf("Response: %s", resp.Message)
		}
		// 减少请求间隔，加快故障触发
		time.Sleep(100 * time.Millisecond)
	}

	// 等待一下，让服务器恢复
	fmt.Println("\n* Waiting for Server A to recover...")
	time.Sleep(12 * time.Second)

	// 验证恢复后是否正常
	fmt.Println("\n* Verifying calls after server recovery")
	for i := 0; i < 3; i++ {
		resp, err := client.SayHello(ctx, fmt.Sprintf("RecoveryUser%d", i), "Hello")
		if err != nil {
			log.Printf("Error: %v", err)
		} else {
			log.Printf("Response: %s", resp.Message)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// 测试熔断器
func testCircuitBreaker(ctx context.Context, client *GreetClient) {
	fmt.Println("\n=== Circuit Breaker Test ===")
	fmt.Println("* Sending a large number of requests to a specific server to trigger circuit breaking")

	// 创建一个短超时的上下文，强制产生超时错误以触发熔断
	timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()

	// 尝试多次请求，触发熔断
	for i := 0; i < 5; i++ {
		_, err := client.SayHello(timeoutCtx, "CircuitBreakerTest", "Hello")
		log.Printf("Circuit breaker test request #%d result: %v", i, err)

		// 使用正常上下文再发一次请求，看是否被熔断
		if i >= 3 {
			_, err = client.SayHello(ctx, "CircuitBreakerTest", "Hello")
			log.Printf("Normal context request #%d result: %v", i, err)
		}

		time.Sleep(300 * time.Millisecond)
	}

	// 等待熔断恢复
	fmt.Println("\n* Waiting for circuit breaker reset...")
	time.Sleep(6 * time.Second)

	// 验证熔断恢复
	resp, err := client.SayHello(ctx, "CircuitBreakerRecovery", "Hello")
	if err != nil {
		log.Printf("Circuit breaker recovery test error: %v", err)
	} else {
		log.Printf("Circuit breaker recovery test success: %s", resp.Message)
	}
}

// 测试并发调用
func testConcurrentCalls(ctx context.Context, client *GreetClient) {
	fmt.Println("\n=== Concurrent Call Test ===")
	fmt.Println("* Initiating multiple concurrent requests")

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			resp, err := client.SayHello(ctx, fmt.Sprintf("ConcurrentUser%d", id), "Hello")
			if err != nil {
				log.Printf("Concurrent call #%d error: %v", id, err)
			} else {
				log.Printf("Concurrent call #%d success: %s", id, resp.Message)
			}
		}(i)

		// 稍微错开并发请求的开始时间
		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()
}

// 显示故障转移指标
func showFailoverMetrics(client *GreetClient) {
	metrics := client.GetFailoverMetrics()
	if metrics == nil {
		fmt.Println("\nUnable to retrieve failover metrics")
		return
	}

	total, success, failure := metrics.GetRequestCount()
	retryCount := metrics.GetRetryCount()
	failoverCount := metrics.GetFailoverCount()
	circuitBreaks := metrics.GetCircuitBreaks()
	avgResponseTime := metrics.GetAvgResponseTime()

	fmt.Println("\n=== Failover Metrics ===")
	fmt.Printf("Total requests: %d (Success: %d, Failure: %d)\n", total, success, failure)
	fmt.Printf("Retry count: %d\n", retryCount)
	fmt.Printf("Failover count: %d\n", failoverCount)
	fmt.Printf("Circuit break count: %d\n", circuitBreaks)
	fmt.Printf("Average response time: %v\n", avgResponseTime)

	// 显示事件日志
	events := metrics.GetEvents()
	if len(events) > 0 {
		fmt.Println("\nRecent events:")
		for i, event := range events {
			if i >= 10 {
				fmt.Println("... (More events omitted)")
				break
			}
			fmt.Printf("  %s\n", event)
		}
	}
}
