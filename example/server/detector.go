package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fyerfyer/fyer-rpc/example/common"
)

// HealthDetector 健康检测器，管理服务健康状态并模拟故障情况
type HealthDetector struct {
	config        *common.ServerConfig
	requestCount  int64  // 请求计数器
	failureTime   *int64 // 故障开始时间（如果当前处于故障状态）
	statusHandler http.Handler
	metrics       *common.SimpleMetrics
	mu            sync.RWMutex
}

// HealthStatus 健康状态响应
type HealthStatus struct {
	Status       string `json:"status"`
	InstanceID   string `json:"instance_id"`
	RequestCount int64  `json:"request_count"`
	Message      string `json:"message,omitempty"`
	Timestamp    int64  `json:"timestamp"`
}

// NewHealthDetector 创建一个新的健康检测器
func NewHealthDetector(config *common.ServerConfig, metrics *common.SimpleMetrics) *HealthDetector {
	detector := &HealthDetector{
		config:      config,
		failureTime: new(int64),
		metrics:     metrics,
	}

	// 创建HTTP处理器
	mux := http.NewServeMux()
	mux.HandleFunc("/health", detector.handleHealth)
	mux.HandleFunc("/metrics", detector.handleMetrics)
	mux.HandleFunc("/status", detector.handleStatus)
	detector.statusHandler = mux

	return detector
}

// Start 启动健康检测HTTP服务
func (d *HealthDetector) Start(port string) {
	go func() {
		log.Printf("Starting health detector on port %s", port)
		if err := http.ListenAndServe(port, d.statusHandler); err != nil {
			log.Printf("Health detector error: %v", err)
		}
	}()
}

// IncrementRequestCount 增加请求计数
func (d *HealthDetector) IncrementRequestCount() int64 {
	return atomic.AddInt64(&d.requestCount, 1)
}

// IsHealthy 检查服务是否健康
func (d *HealthDetector) IsHealthy() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 如果设置了随机故障率，根据概率返回失败
	if d.config.FailRate > 0 {
		if rand.Float64() < d.config.FailRate {
			return false
		}
	}

	// 如果设置了处理特定数量请求后故障
	if d.config.FailAfter > 0 {
		count := atomic.LoadInt64(&d.requestCount)
		if count >= int64(d.config.FailAfter) {
			failTime := atomic.LoadInt64(d.failureTime)

			// 如果这是第一次触发故障，记录故障时间
			if failTime == 0 {
				atomic.StoreInt64(d.failureTime, time.Now().UnixNano())
				return false
			}

			// 如果故障持续时间已过，恢复服务
			if d.config.FailDuration > 0 {
				failureTimestamp := time.Unix(0, failTime)
				if time.Since(failureTimestamp) > d.config.FailDuration {
					// 重置故障时间，表示已恢复
					atomic.StoreInt64(d.failureTime, 0)
					// 也可以选择重置计数器，取决于需求
					// atomic.StoreInt64(&d.requestCount, 0)
					return true
				}
			}

			return false
		}
	}

	return true
}

// handleHealth 处理健康检查请求
func (d *HealthDetector) handleHealth(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 增加请求计数
	d.IncrementRequestCount()

	// 检查健康状态
	isHealthy := d.IsHealthy()

	// 构建响应
	status := HealthStatus{
		Status:       "UP",
		InstanceID:   d.config.ID,
		RequestCount: atomic.LoadInt64(&d.requestCount),
		Timestamp:    time.Now().Unix(),
	}

	if !isHealthy {
		status.Status = "DOWN"
		status.Message = "Service is currently unavailable"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	// 发送JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)

	// 记录指标
	d.metrics.RecordRequest(r.RemoteAddr, time.Since(start), nil)
}

// handleMetrics 处理指标请求
func (d *HealthDetector) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// 获取当前的指标数据
	total, success, failure := d.metrics.GetRequestCount()
	retries := d.metrics.GetRetryCount()
	failovers := d.metrics.GetFailoverCount()
	circuitBreaks := d.metrics.GetCircuitBreaks()
	avgResponse := d.metrics.GetAvgResponseTime()

	// 构建响应
	metrics := struct {
		TotalRequests   int64         `json:"total_requests"`
		SuccessCount    int64         `json:"success_count"`
		FailureCount    int64         `json:"failure_count"`
		RetryCount      int64         `json:"retry_count"`
		FailoverCount   int64         `json:"failover_count"`
		CircuitBreaks   int64         `json:"circuit_breaks"`
		AvgResponseTime time.Duration `json:"avg_response_time"`
		InstanceID      string        `json:"instance_id"`
		InstanceStatus  string        `json:"instance_status"`
		Timestamp       int64         `json:"timestamp"`
	}{
		TotalRequests:   total,
		SuccessCount:    success,
		FailureCount:    failure,
		RetryCount:      retries,
		FailoverCount:   failovers,
		CircuitBreaks:   circuitBreaks,
		AvgResponseTime: avgResponse,
		InstanceID:      d.config.ID,
		InstanceStatus:  "UP",
		Timestamp:       time.Now().Unix(),
	}

	if !d.IsHealthy() {
		metrics.InstanceStatus = "DOWN"
	}

	// 发送JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleStatus 处理服务状态请求，包含详细状态信息
func (d *HealthDetector) handleStatus(w http.ResponseWriter, r *http.Request) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 构建详细的状态信息
	status := struct {
		InstanceID    string                          `json:"instance_id"`
		Address       string                          `json:"address"`
		Port          int                             `json:"port"`
		Status        string                          `json:"status"`
		RequestCount  int64                           `json:"request_count"`
		FailAfter     int                             `json:"fail_after"`
		FailDuration  time.Duration                   `json:"fail_duration"`
		FailRate      float64                         `json:"fail_rate"`
		FailureTime   *time.Time                      `json:"failure_time,omitempty"`
		InstanceStats map[string]*common.InstanceStat `json:"instance_stats,omitempty"`
		Events        []string                        `json:"events,omitempty"`
		Timestamp     int64                           `json:"timestamp"`
	}{
		InstanceID:    d.config.ID,
		Address:       d.config.Address,
		Port:          d.config.Port,
		Status:        "UP",
		RequestCount:  atomic.LoadInt64(&d.requestCount),
		FailAfter:     d.config.FailAfter,
		FailDuration:  d.config.FailDuration,
		FailRate:      d.config.FailRate,
		InstanceStats: d.metrics.GetInstanceStats(),
		Events:        d.metrics.GetEvents(),
		Timestamp:     time.Now().Unix(),
	}

	// 如果处于故障状态，计算并填充故障时间信息
	failTime := atomic.LoadInt64(d.failureTime)
	if failTime > 0 {
		failureTime := time.Unix(0, failTime)
		status.FailureTime = &failureTime
		status.Status = "DOWN"
	} else if !d.IsHealthy() {
		status.Status = "DOWN"
	}

	// 发送JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// Reset 重置检测器状态
func (d *HealthDetector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	atomic.StoreInt64(&d.requestCount, 0)
	atomic.StoreInt64(d.failureTime, 0)
}
