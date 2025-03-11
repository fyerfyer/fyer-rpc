package failover

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
)

// DefaultFailoverHandler 默认故障转移处理器
type DefaultFailoverHandler struct {
	detector        Detector
	circuitBreaker  CircuitBreaker
	retryPolicy     RetryPolicy
	recovery        RecoveryStrategy
	monitor         InstanceMonitor
	config          *Config
	instanceManager *InstanceManager
	mu              sync.RWMutex
}

// InstanceManager 实例管理器，用于在故障时选择下一个可用实例
type InstanceManager struct {
	instances    []*naming.Instance
	statusMap    map[string]Status
	nextIndexMap map[string]int
	mu           sync.RWMutex
}

// NewInstanceManager 创建实例管理器
func NewInstanceManager(instances []*naming.Instance) *InstanceManager {
	manager := &InstanceManager{
		instances:    instances,
		statusMap:    make(map[string]Status),
		nextIndexMap: make(map[string]int),
	}

	// 初始化所有实例状态为健康
	for _, instance := range instances {
		manager.statusMap[instance.ID] = StatusHealthy
	}

	return manager
}

// UpdateInstances 更新实例列表
func (m *InstanceManager) UpdateInstances(instances []*naming.Instance) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 保存旧的状态
	oldStatusMap := m.statusMap

	// 创建新的状态映射
	newStatusMap := make(map[string]Status)
	for _, instance := range instances {
		if status, ok := oldStatusMap[instance.ID]; ok {
			newStatusMap[instance.ID] = status
		} else {
			newStatusMap[instance.ID] = StatusHealthy
		}
	}

	// 更新实例列表和状态映射
	m.instances = instances
	m.statusMap = newStatusMap
}

// MarkInstanceStatus 标记实例状态
func (m *InstanceManager) MarkInstanceStatus(instanceID string, status Status) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusMap[instanceID] = status
}

// GetInstance 获取一个可用的实例
func (m *InstanceManager) GetInstance(strategy string) (*naming.Instance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.instances) == 0 {
		return nil, ErrNoAvailableInstances
	}

	// 统计健康实例数量
	var healthyInstances []*naming.Instance
	for _, instance := range m.instances {
		if status, ok := m.statusMap[instance.ID]; !ok || (status == StatusHealthy || status == StatusSuspect) {
			healthyInstances = append(healthyInstances, instance)
		}
	}

	if len(healthyInstances) == 0 {
		return nil, ErrNoAvailableInstances
	}

	// 根据策略选择实例
	switch strategy {
	case "next":
		return m.nextInstance(), nil
	case "random":
		return m.randomInstance(healthyInstances), nil
	case "best":
		return m.bestInstance(healthyInstances), nil
	default:
		return m.nextInstance(), nil
	}
}

// nextInstance 轮询选择下一个实例
func (m *InstanceManager) nextInstance() *naming.Instance {
	serviceKey := "default"
	if len(m.instances) == 0 {
		return nil
	}

	// 获取当前索引
	index, ok := m.nextIndexMap[serviceKey]
	if !ok {
		index = 0
	}

	// 最多尝试所有实例
	for i := 0; i < len(m.instances); i++ {
		// 计算下一个索引
		nextIndex := (index + i) % len(m.instances)
		instance := m.instances[nextIndex]

		// 检查实例是否可用
		if status, ok := m.statusMap[instance.ID]; !ok || (status == StatusHealthy || status == StatusSuspect) {
			// 更新下一个索引
			m.nextIndexMap[serviceKey] = (nextIndex + 1) % len(m.instances)
			return instance
		}
	}

	// 如果没有健康实例，返回随机一个
	randIndex := index % len(m.instances)
	m.nextIndexMap[serviceKey] = (randIndex + 1) % len(m.instances)
	return m.instances[randIndex]
}

// randomInstance 随机选择一个实例
func (m *InstanceManager) randomInstance(healthyInstances []*naming.Instance) *naming.Instance {
	if len(healthyInstances) == 0 {
		return nil
	}
	return healthyInstances[rand.Intn(len(healthyInstances))]
}

// bestInstance 选择"最优"实例（这里简单实现为选择第一个健康实例）
func (m *InstanceManager) bestInstance(healthyInstances []*naming.Instance) *naming.Instance {
	if len(healthyInstances) == 0 {
		return nil
	}
	return healthyInstances[0]
}

// NewFailoverHandler 创建故障转移处理器
func NewFailoverHandler(config *Config) (*DefaultFailoverHandler, error) {
	// 创建各组件
	detector := NewDetector(config, "timeout") // 默认使用超时检测器
	circuitBreaker := NewCircuitBreaker(config)
	retryPolicy := NewRetryPolicy(config)
	recovery := NewRecoveryStrategy(config, detector)
	monitor := NewInstanceMonitor(config)

	return &DefaultFailoverHandler{
		detector:        detector,
		circuitBreaker:  circuitBreaker,
		retryPolicy:     retryPolicy,
		recovery:        recovery,
		monitor:         monitor,
		config:          config,
		instanceManager: NewInstanceManager(nil), // 先创建空的实例管理器
	}, nil
}

// Execute 执行带故障转移的调用
func (h *DefaultFailoverHandler) Execute(ctx context.Context, instances []*naming.Instance, operation func(context.Context,
	*naming.Instance) error) (*FailoverResult, error) {
	// 检查上下文是否已取消
	if ctx.Err() != nil {
		return &FailoverResult{
			Success:     false,
			RetryCount:  0,
			FailedNodes: make([]string, 0),
			// 设置一个默认实例，避免返回nil
			Instance: instances[0],
		}, ctx.Err()
	}

	h.mu.RLock()
	// 更新实例列表
	h.instanceManager.UpdateInstances(instances)
	h.mu.RUnlock()

	// 创建用于记录结果的对象
	result := &FailoverResult{
		RetryCount:  0,
		FailedNodes: make([]string, 0),
	}

	// 检查实例列表是否为空
	if len(instances) == 0 {
		return result, ErrNoAvailableInstances
	}

	// 记录开始时间
	startTime := time.Now()

	// 获取故障转移策略
	failoverStrategy := h.config.FailoverStrategy

	// 记录错误
	var lastErr error
	circuitBreakerTriggered := false

	// 记录已尝试过的实例ID，避免重复尝试
	triedInstances := make(map[string]bool)

	// 使用重试策略
	maxRetries := h.retryPolicy.MaxAttempts()

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			result.RetryCount++

			// 检查是否需要继续重试
			if !h.retryPolicy.ShouldRetry(ctx, attempt, lastErr) {
				break
			}

			// 计算退避时间
			backoff := h.retryPolicy.NextBackoff(attempt)

			// 创建定时器
			timer := time.NewTimer(backoff)

			// 等待退避时间或上下文取消
			select {
			case <-ctx.Done():
				timer.Stop()
				result.Duration = time.Since(startTime)
				return result, ctx.Err()
			case <-timer.C:
				// 继续下一次尝试
			}
		}

		// 选择一个实例
		instance, err := h.instanceManager.GetInstance(failoverStrategy)
		if err != nil {
			lastErr = err
			break // 没有可用实例，直接中断循环
		}

		// 如果已经尝试过这个实例，尝试再获取一个不同的实例
		if triedInstances[instance.ID] {
			// 如果所有实例都已尝试过，则退出循环
			if len(triedInstances) >= len(instances) {
				break
			}
			// 标记实例为不可用，并尝试获取下一个
			h.instanceManager.MarkInstanceStatus(instance.ID, StatusUnhealthy)
			continue
		}

		// 标记已尝试
		triedInstances[instance.ID] = true

		// 检查熔断器状态
		allow, cbErr := h.circuitBreaker.Allow(ctx, instance)
		if cbErr != nil || !allow {
			// 记录失败节点
			result.FailedNodes = append(result.FailedNodes, instance.Address)

			if cbErr == ErrCircuitOpen || !allow {
				circuitBreakerTriggered = true
				lastErr = ErrCircuitOpen
			} else if cbErr != nil {
				lastErr = cbErr
			}
			continue
		}

		// 检测实例健康状态
		status, err := h.detector.Detect(ctx, instance)
		if err != nil || status != StatusHealthy {
			h.handleFailure(ctx, instance, err)
			result.FailedNodes = append(result.FailedNodes, instance.Address)
			if err != nil {
				lastErr = err
			} else {
				lastErr = ErrServiceUnavailable
			}
			continue
		}

		// 执行操作
		operationStartTime := time.Now()
		err = operation(ctx, instance)
		operationDuration := time.Since(operationStartTime)

		if err != nil {
			// 操作失败，标记失败，更新指标，记录错误
			h.handleFailure(ctx, instance, err)
			result.FailedNodes = append(result.FailedNodes, instance.Address)
			lastErr = err
			continue
		}

		// 操作成功，标记成功，更新指标
		h.handleSuccess(ctx, instance, operationDuration)
		result.Success = true
		result.Instance = instance
		result.Duration = time.Since(startTime)

		return result, nil
	}

	// 设置最终结果
	result.Duration = time.Since(startTime)
	result.Error = lastErr

	// 设置最后一个实例（即使失败也需要设置）
	for _, inst := range instances {
		if !triedInstances[inst.ID] {
			result.Instance = inst
			break
		}
	}

	// 如果没有设置实例，设置第一个实例作为默认
	if result.Instance == nil && len(instances) > 0 {
		result.Instance = instances[0]
	}

	// 如果有成功的尝试（说明最终调用成功），则应该返回成功
	if result.Success {
		return result, nil
	}

	// 如果已触发熔断器，优先返回熔断器错误
	if circuitBreakerTriggered {
		return result, ErrCircuitOpen
	}

	// 如果已达到或超过最大重试次数，则返回特定错误
	if result.RetryCount >= maxRetries {
		return result, ErrMaxRetriesExceeded
	}

	return result, lastErr
}

// handleSuccess 处理成功调用
func (h *DefaultFailoverHandler) handleSuccess(ctx context.Context, instance *naming.Instance, duration time.Duration) {
	// 标记实例为成功状态
	h.detector.MarkSuccess(ctx, instance)

	// 标记熔断器成功
	h.circuitBreaker.MarkSuccess(ctx, instance)

	// 更新监控指标
	h.monitor.ReportSuccess(ctx, instance, duration)

	// 更新实例管理器中的实例状态
	h.instanceManager.MarkInstanceStatus(instance.ID, StatusHealthy)
}

// handleFailure 处理失败调用
func (h *DefaultFailoverHandler) handleFailure(ctx context.Context, instance *naming.Instance, err error) {
	// 标记实例为失败状态
	h.detector.MarkFailed(ctx, instance)

	// 标记熔断器失败
	h.circuitBreaker.MarkFailure(ctx, instance, err)

	// 更新监控指标
	h.monitor.ReportFailure(ctx, instance, err)

	// 更新实例管理器中的实例状态
	h.instanceManager.MarkInstanceStatus(instance.ID, StatusUnhealthy)
}

// GetDetector 获取故障检测器
func (h *DefaultFailoverHandler) GetDetector() Detector {
	return h.detector
}

// GetCircuitBreaker 获取熔断器
func (h *DefaultFailoverHandler) GetCircuitBreaker() CircuitBreaker {
	return h.circuitBreaker
}

// GetRetryPolicy 获取重试策略
func (h *DefaultFailoverHandler) GetRetryPolicy() RetryPolicy {
	return h.retryPolicy
}

// GetRecoveryStrategy 获取恢复策略
func (h *DefaultFailoverHandler) GetRecoveryStrategy() RecoveryStrategy {
	return h.recovery
}

// GetMonitor 获取实例监控器
func (h *DefaultFailoverHandler) GetMonitor() InstanceMonitor {
	return h.monitor
}

// UpdateConfig 更新配置
func (h *DefaultFailoverHandler) UpdateConfig(config *Config) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = config
}

// NoopCircuitBreaker 空操作熔断器，总是允许请求通过
type NoopCircuitBreaker struct{}

func (c *NoopCircuitBreaker) Allow(ctx context.Context, instance *naming.Instance) (bool, error) {
	return true, nil
}

func (c *NoopCircuitBreaker) MarkSuccess(ctx context.Context, instance *naming.Instance) error {
	return nil
}

func (c *NoopCircuitBreaker) MarkFailure(ctx context.Context, instance *naming.Instance, err error) error {
	return nil
}

func (c *NoopCircuitBreaker) GetState(instance *naming.Instance) (State, error) {
	return StateClosed, nil
}

func (c *NoopCircuitBreaker) Reset(instance *naming.Instance) error {
	return nil
}
