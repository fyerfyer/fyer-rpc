package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

func main() {
	// 创建客户端
	client, err := NewGreetClient("GreetService")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 创建上下文
	ctx := context.Background()

	// 测试并发调用
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			testClient(ctx, client, id)
		}(i)
	}

	// 等待所有调用完成
	wg.Wait()

	// 添加一个延时，让 Prometheus 有足够时间抓取指标
	fmt.Println("Service test completed. Waiting 100 seconds for metrics collection...")
	time.Sleep(100 * time.Second)

	// 获取并打印统计信息
	stats, err := client.GetStats(ctx)
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
		return
	}

	fmt.Printf("\nService Statistics:\n")
	fmt.Printf("Total Greets: %d\n", stats.TotalGreets)
	fmt.Printf("Last Greet Time: %s\n", time.Unix(0, stats.LastGreetTime).Format(time.RFC3339))
	fmt.Printf("Greets by Name:\n")
	for name, count := range stats.GreetsByName {
		fmt.Printf("  %s: %d\n", name, count)
	}
}

func testClient(ctx context.Context, client *GreetClient, id int) {
	names := []string{"Alice", "Bob", "Charlie", "David", "Eve"}
	greetings := []string{"Hello", "Hi", "Hey", "Greetings", ""}

	// 每个goroutine发送多次请求
	for j := 0; j < 5; j++ {
		name := names[j%len(names)]
		greeting := greetings[j%len(greetings)]

		resp, err := client.SayHello(ctx, name, greeting)
		if err != nil {
			log.Printf("Error in goroutine %d: %v", id, err)
			continue
		}

		log.Printf("Goroutine %d received: %s (time: %s)",
			id,
			resp.Message,
			time.Unix(0, resp.GreetTime).Format(time.RFC3339Nano))

		// 随机休眠一段时间，模拟真实场景
		time.Sleep(time.Duration(100+id*50) * time.Millisecond)
	}
}
