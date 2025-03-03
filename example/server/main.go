package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fyerfyer/fyer-rpc/example/common"
	"github.com/fyerfyer/fyer-rpc/registry/etcd"
	"github.com/fyerfyer/fyer-rpc/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 命令行参数
var (
	port         = flag.Int("port", 8001, "服务器端口")
	id           = flag.String("id", "server-A", "服务器ID")
	failAfter    = flag.Int("fail-after", 0, "在处理这么多个请求后故障")
	failDuration = flag.Duration("fail-duration", 0, "故障持续时间")
	failRate     = flag.Float64("fail-rate", 0.0, "随机故障概率 (0-1)")
)

// Prometheus指标
var (
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fyerrpc_request_total",
			Help: "The total number of processed requests",
		},
		[]string{"instance", "status"},
	)

	responseTimeHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fyerrpc_response_time_seconds",
			Help:    "Response time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"instance", "method"},
	)

	failoverCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "fyerrpc_failover_total",
			Help: "The total number of failover events",
		},
	)

	circuitBreakerCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fyerrpc_circuit_breaks",
			Help: "The total number of circuit breaker state changes",
		},
		[]string{"instance", "state"},
	)

	instanceHealthGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fyerrpc_instance_health",
			Help: "Instance health status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"instance"},
	)
)

func main() {
	// 解析命令行参数
	flag.Parse()

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

	// 根据命令行参数创建服务器配置
	serverConfig := &common.ServerConfig{
		ID:           *id,
		Address:      "localhost",
		Port:         *port,
		FailAfter:    *failAfter,
		FailDuration: *failDuration,
		FailRate:     *failRate,
	}

	// 创建全局的 Prometheus 指标服务
	metricsHandler := http.NewServeMux()
	metricsHandler.Handle("/metrics", promhttp.Handler())
	go func() {
		metricsPort := 8081
		log.Printf("Starting metrics server on port %d", metricsPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), metricsHandler); err != nil {
			log.Printf("Metrics server failed: %v", err)
		}
	}()

	// 创建全局指标收集器
	globalMetrics := common.NewSimpleMetrics(500, 100)

	// 将SimpleMetrics集成到Prometheus
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C
			total, success, failure := globalMetrics.GetRequestCount()
			log.Printf("Metrics update: total=%d, success=%d, failure=%d", total, success, failure)

			// 更新请求计数
			requestCounter.WithLabelValues(serverConfig.ID, "success").Add(float64(success))
			requestCounter.WithLabelValues(serverConfig.ID, "failure").Add(float64(failure))

			// 实例健康状态指标
			instanceStats := globalMetrics.GetInstanceStats()
			for addr, stat := range instanceStats {
				var healthStatus float64 = 1.0
				if stat.CircuitState == "open" {
					healthStatus = 0.0
				}
				instanceHealthGauge.WithLabelValues(addr).Set(healthStatus)

				// 响应时间指标
				if stat.AvgLatency > 0 {
					responseTimeHistogram.WithLabelValues(serverConfig.ID, "SayHello").Observe(stat.AvgLatency.Seconds())
				}
			}
		}
	}()

	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动单个服务器实例
	var wg sync.WaitGroup
	wg.Add(1)

	// 为服务器创建独立的 RPC 服务器
	server := rpc.NewServer()

	// 创建并启动 Greet 服务
	address := fmt.Sprintf("%s:%d", serverConfig.Address, serverConfig.Port)
	greetServer := NewGreetServer(registry, serverConfig)

	// 使用全局指标收集器
	greetServer.metrics = globalMetrics

	if err := server.RegisterService(greetServer.greetService); err != nil {
		log.Printf("Failed to register service %s: %v", serverConfig.ID, err)
		return
	}

	// 启动服务注册
	if err := greetServer.Start(ctx); err != nil {
		log.Printf("Failed to start service %s: %v", serverConfig.ID, err)
		return
	}

	// 启动 RPC 服务器
	log.Printf("Starting RPC server %s at %s", serverConfig.ID, address)
	go func() {
		defer wg.Done()
		if err := server.Start(address); err != nil {
			log.Printf("RPC server %s failed: %v", serverConfig.ID, err)
		}
	}()

	// 显示服务器的健康检查URL
	healthPort := serverConfig.Port + 10000
	log.Printf("Server started. Health check endpoint: http://%s:%d/health",
		serverConfig.Address, healthPort)

	// 等待终止信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// 优雅关闭
	log.Println("Shutting down...")
	cancel() // 通知所有服务器关闭

	// 等待所有服务器关闭，设置超时时间
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// 注销服务
	if err := greetServer.Stop(shutdownCtx); err != nil {
		log.Printf("Failed to stop service %s: %v", greetServer.config.ID, err)
	}

	// 等待所有服务器线程完成
	wg.Wait()
	log.Println("Server stopped")
}
