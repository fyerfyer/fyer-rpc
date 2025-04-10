package failover

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
)

// BaseDetector 基础故障检测器
type BaseDetector struct {
	statusMap    map[string]Status    // 实例ID -> 状态
	failureCount map[string]int       // 实例失败计数
	successCount map[string]int       // 实例成功计数
	timestamps   map[string]time.Time // 记录实例状态最后更新时间
	config       *Config              // 故障检测配置
	mu           sync.RWMutex         // 保护共享数据的互斥锁
}

// NewBaseDetector 创建基础检测器
func NewBaseDetector(config *Config) *BaseDetector {
	return &BaseDetector{
		statusMap:    make(map[string]Status),
		failureCount: make(map[string]int),
		successCount: make(map[string]int),
		timestamps:   make(map[string]time.Time),
		config:       config,
	}
}

// Detect 检测实例是否健康，基础实现返回已记录的状态
func (d *BaseDetector) Detect(ctx context.Context, instance *naming.Instance) (Status, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// 获取已记录的状态，如果没有则默认为健康
	status, ok := d.statusMap[instance.ID]
	if !ok {
		return StatusHealthy, nil
	}

	return status, nil
}

// MarkFailed 标记实例为失败状态
func (d *BaseDetector) MarkFailed(ctx context.Context, instance *naming.Instance) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 增加失败计数
	d.failureCount[instance.ID]++
	d.successCount[instance.ID] = 0 // 重置成功计数
	d.timestamps[instance.ID] = time.Now()

	// 判断是否达到失败阈值
	if d.failureCount[instance.ID] >= d.config.FailureThreshold {
		d.statusMap[instance.ID] = StatusUnhealthy
	} else {
		d.statusMap[instance.ID] = StatusSuspect
	}

	return nil
}

// MarkSuccess 标记实例为成功状态
func (d *BaseDetector) MarkSuccess(ctx context.Context, instance *naming.Instance) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 增加成功计数
	d.successCount[instance.ID]++
	d.failureCount[instance.ID] = 0 // 重置失败计数
	d.timestamps[instance.ID] = time.Now()

	// 判断是否达到成功阈值
	if d.successCount[instance.ID] >= d.config.SuccessThreshold {
		d.statusMap[instance.ID] = StatusHealthy
	}

	return nil
}

// GetStatus 获取实例状态
func (d *BaseDetector) GetStatus(instance *naming.Instance) Status {
	d.mu.RLock()
	defer d.mu.RUnlock()

	status, ok := d.statusMap[instance.ID]
	if !ok {
		return StatusHealthy // 默认为健康状态
	}
	return status
}

// TimeoutDetector 基于超时的故障检测器
type TimeoutDetector struct {
	*BaseDetector
}

// NewTimeoutDetector 创建基于超时的故障检测器
func NewTimeoutDetector(config *Config) *TimeoutDetector {
	return &TimeoutDetector{
		BaseDetector: NewBaseDetector(config),
	}
}

// Detect 通过连接超时判断实例健康状态
func (d *TimeoutDetector) Detect(ctx context.Context, instance *naming.Instance) (Status, error) {
	// 首先获取已记录的状态
	status, _ := d.BaseDetector.Detect(ctx, instance)

	// 如果实例已被标记为不健康，直接返回
	if status == StatusUnhealthy || status == StatusIsolated {
		return status, nil
	}

	// 设置超时上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, d.config.ConnectionTimeout)
	defer cancel()

	// 尝试连接实例
	var dialer net.Dialer
	conn, err := dialer.DialContext(timeoutCtx, "tcp", instance.Address)
	if err != nil {
		// 连接失败，标记失败
		d.MarkFailed(ctx, instance)
		return d.GetStatus(instance), err
	}
	conn.Close()

	// 连接成功，标记为成功
	d.MarkSuccess(ctx, instance)
	return d.GetStatus(instance), nil
}

// ErrorRateDetector 基于错误率的故障检测器
type ErrorRateDetector struct {
	*BaseDetector
	requestCount map[string]int    // 总请求计数
	errorCount   map[string]int    // 错误请求计数
	windows      map[string][]bool // 滑动窗口，记录最近的请求成功/失败
	windowSize   int               // 滑动窗口大小
}

// NewErrorRateDetector 创建基于错误率的故障检测器
func NewErrorRateDetector(config *Config, windowSize int) *ErrorRateDetector {
	if windowSize <= 0 {
		windowSize = 100 // 默认窗口大小
	}

	return &ErrorRateDetector{
		BaseDetector: NewBaseDetector(config),
		requestCount: make(map[string]int),
		errorCount:   make(map[string]int),
		windows:      make(map[string][]bool),
		windowSize:   windowSize,
	}
}

// ReportRequest 报告请求结果
func (d *ErrorRateDetector) ReportRequest(ctx context.Context, instance *naming.Instance, success bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	instanceID := instance.ID

	// 初始化窗口
	if _, ok := d.windows[instanceID]; !ok {
		d.windows[instanceID] = make([]bool, 0, d.windowSize)
	}

	// 更新滑动窗口
	window := d.windows[instanceID]
	if len(window) >= d.windowSize {
		// 移除最早的记录
		oldSuccess := window[0]
		window = window[1:]
		if !oldSuccess {
			d.errorCount[instanceID]--
		}
		d.requestCount[instanceID]--
	}

	// 添加新的记录
	window = append(window, success)
	d.windows[instanceID] = window
	d.requestCount[instanceID]++
	if !success {
		d.errorCount[instanceID]++
	}

	// 计算错误率
	errorRate := float64(d.errorCount[instanceID]) / float64(d.requestCount[instanceID])

	// 根据错误率更新状态
	if errorRate > 0.5 && d.requestCount[instanceID] >= 10 { // 错误率超过50%且有足够样本
		d.statusMap[instanceID] = StatusUnhealthy
	} else if errorRate > 0.2 && d.requestCount[instanceID] >= 5 { // 错误率超过20%且有一定样本
		d.statusMap[instanceID] = StatusSuspect
	} else if d.requestCount[instanceID] >= 5 { // 有足够的成功样本
		d.statusMap[instanceID] = StatusHealthy
	}

	// 更新时间戳
	d.timestamps[instanceID] = time.Now()
}

// Detect 检测实例健康状态
func (d *ErrorRateDetector) Detect(ctx context.Context, instance *naming.Instance) (Status, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	status, ok := d.statusMap[instance.ID]
	if !ok {
		return StatusHealthy, nil
	}

	// 检查是否需要重置状态
	lastUpdate, ok := d.timestamps[instance.ID]
	if ok && time.Since(lastUpdate) > d.config.FailureDetectionTime {
		// 状态过期，重置为健康状态
		return StatusHealthy, nil
	}

	return status, nil
}

// HealthCheckDetector 健康检查故障检测器
type HealthCheckDetector struct {
	*BaseDetector
	checker        func(context.Context, *naming.Instance) (bool, error) // 健康检查函数
	checkInterval  time.Duration                                         // 检查间隔
	stopCh         chan struct{}                                         // 停止信号
	instancesCache []*naming.Instance                                    // 检查的实例缓存
	running        bool                                                  // 是否正在运行
	runMutex       sync.Mutex                                            // 保护running字段的互斥锁
}

// NewHealthCheckDetector 创建健康检查故障检测器
func NewHealthCheckDetector(config *Config, checker func(context.Context, *naming.Instance) (bool, error)) *HealthCheckDetector {
	return &HealthCheckDetector{
		BaseDetector:   NewBaseDetector(config),
		checker:        checker,
		checkInterval:  config.FailureDetectionTime,
		stopCh:         make(chan struct{}),
		instancesCache: make([]*naming.Instance, 0),
	}
}

// Start 启动健康检查
func (d *HealthCheckDetector) Start() {
	d.runMutex.Lock()
	defer d.runMutex.Unlock()

	if d.running {
		return
	}

	d.running = true
	go d.checkLoop()
}

// Stop 停止健康检查
func (d *HealthCheckDetector) Stop() {
	d.runMutex.Lock()
	defer d.runMutex.Unlock()

	if !d.running {
		return
	}

	close(d.stopCh)
	d.running = false
}

// UpdateInstances 更新要检查的实例列表
func (d *HealthCheckDetector) UpdateInstances(instances []*naming.Instance) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.instancesCache = make([]*naming.Instance, len(instances))
	copy(d.instancesCache, instances)
}

// checkLoop 健康检查循环
func (d *HealthCheckDetector) checkLoop() {
	ticker := time.NewTicker(d.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.checkInstances()
		}
	}
}

// checkInstances 检查所有实例的健康状态
func (d *HealthCheckDetector) checkInstances() {
	d.mu.RLock()
	instances := d.instancesCache
	d.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), d.config.RequestTimeout)
	defer cancel()

	for _, instance := range instances {
		healthy, _ := d.checker(ctx, instance)
		if healthy {
			d.MarkSuccess(ctx, instance)
		} else {
			d.MarkFailed(ctx, instance)
		}
	}
}

// Detect 检测实例健康状态
func (d *HealthCheckDetector) Detect(ctx context.Context, instance *naming.Instance) (Status, error) {
	// 首先获取缓存的状态
	status, _ := d.BaseDetector.Detect(ctx, instance)

	// 如果实例已被标记为不健康或隔离，直接返回
	if status == StatusUnhealthy || status == StatusIsolated {
		return status, nil
	}

	// 进行实时健康检查
	healthy, err := d.checker(ctx, instance)
	if err != nil || !healthy {
		d.MarkFailed(ctx, instance)
		return d.GetStatus(instance), err
	}

	d.MarkSuccess(ctx, instance)
	return d.GetStatus(instance), nil
}

// NewDetector 创建检测器的工厂函数
func NewDetector(config *Config, detectorType string) Detector {
	switch detectorType {
	case "timeout":
		return NewTimeoutDetector(config)
	case "error_rate":
		return NewErrorRateDetector(config, 100)
	case "health_check":
		return NewHealthCheckDetector(config, func(ctx context.Context, instance *naming.Instance) (bool, error) {
			// 默认的健康检查实现：尝试建立TCP连接
			var dialer net.Dialer
			conn, err := dialer.DialContext(ctx, "tcp", instance.Address)
			if err != nil {
				return false, err
			}
			conn.Close()
			return true, nil
		})
	default:
		// 默认使用超时检测器
		return NewTimeoutDetector(config)
	}
}
