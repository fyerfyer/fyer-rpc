package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fyerfyer/fyer-rpc/discovery/metrics"
	"github.com/fyerfyer/fyer-rpc/registry/etcd"
	"github.com/fyerfyer/fyer-rpc/rpc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 创建 etcd 注册中心
	registry, err := etcd.New(
		etcd.WithEndpoints([]string{"localhost:2379"}),
		etcd.WithDialTimeout(time.Second*5),
		etcd.WithTTL(10),
	)
	if err != nil {
		log.Fatalf("Failed to create registry: %v", err)
	}
	defer registry.Close()

	// 创建 Prometheus 指标收集器
	// 修改 Prometheus 指标收集器配置
	metricsCollector, err := metrics.NewPrometheusMetrics(&metrics.PrometheusConfig{
		QueryURL: "http://localhost:9090",
		JobName:  "fyerrpc",
	})
	if err != nil {
		log.Fatalf("Failed to create metrics collector: %v", err)
	}
	defer metricsCollector.Close()

	// 创建 RPC 服务器
	server := rpc.NewServer()

	// 创建并启动 Greet 服务
	greetServer := NewGreetServer(registry, "localhost:8080")
	if err := server.RegisterService(greetServer.GreetServiceImpl); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 启动服务器和指标服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务注册
	if err := greetServer.Start(ctx); err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	// 启动 Prometheus 指标服务器
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("Metrics server failed: %v", err)
		}
	}()

	// 启动 RPC 服务器
	go func() {
		if err := server.Start("localhost:8080"); err != nil {
			log.Printf("RPC server failed: %v", err)
		}
	}()

	// 等待终止信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 优雅关闭
	log.Println("Shutting down...")
	cancel()

	// 注销服务
	if err := greetServer.Stop(context.Background()); err != nil {
		log.Printf("Failed to stop service: %v", err)
	}

	log.Println("Server stopped")
}
