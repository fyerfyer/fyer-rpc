package failover

import (
	"context"
	"sync"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
)

// RecoveryManager 恢复管理器，负责管理所有恢复策略
type RecoveryManager struct {
	strategy       RecoveryStrategy
	recoveryStates map[string]*recoveryState // 实例ID -> 恢复状态
	config         *Config
	mu             sync.RWMutex
	done           chan struct{}
}

// recoveryState 实例恢复状态
type recoveryState struct {
	instanceID       string
	lastAttemptTime  time.Time // 上次恢复尝试时间
	attemptCount     int       // 尝试恢复次数
	successCount     int       // 连续成功次数
	instance         *naming.Instance
	status           Status // 当前状态
	recoveryStrategy RecoveryStrategy
	mu               sync.RWMutex
}

// NewRecoveryManager 创建恢复管理器
func NewRecoveryManager(config *Config, strategy RecoveryStrategy) *RecoveryManager {
	manager := &RecoveryManager{
		strategy:       strategy,
		recoveryStates: make(map[string]*recoveryState),
		config:         config,
		done:           make(chan struct{}),
	}

	// 启动恢复任务
	go manager.recoveryLoop()

	return manager
}

// AddInstance 添加需要恢复的实例
func (m *RecoveryManager) AddInstance(ctx context.Context, instance *naming.Instance, status Status) {
	if status != StatusUnhealthy && status != StatusSuspect {
		return // 只处理不健康或可疑实例
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在
	if _, ok := m.recoveryStates[instance.ID]; ok {
		return
	}

	// 添加新的恢复状态
	m.recoveryStates[instance.ID] = &recoveryState{
		instanceID:       instance.ID,
		lastAttemptTime:  time.Time{}, // 未尝试过恢复
		attemptCount:     0,
		successCount:     0,
		instance:         instance,
		status:           status,
		recoveryStrategy: m.strategy,
	}
}

// RemoveInstance 移除实例的恢复状态
func (m *RecoveryManager) RemoveInstance(instanceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.recoveryStates, instanceID)
}

// recoveryLoop 恢复循环，定期检查需要恢复的实例
func (m *RecoveryManager) recoveryLoop() {
	ticker := time.NewTicker(m.config.RecoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.tryRecover()
		}
	}
}

// tryRecover 尝试恢复所有不健康的实例
func (m *RecoveryManager) tryRecover() {
	m.mu.RLock()
	instances := make([]*recoveryState, 0, len(m.recoveryStates))
	for _, state := range m.recoveryStates {
		instances = append(instances, state)
	}
	m.mu.RUnlock()

	for _, state := range instances {
		m.recoverInstance(state)
	}
}

// recoverInstance 尝试恢复单个实例
func (m *RecoveryManager) recoverInstance(state *recoveryState) {
	state.mu.Lock()
	defer state.mu.Unlock()

	// 检查上次恢复尝试的时间间隔
	if time.Since(state.lastAttemptTime) < m.strategy.RecoveryDelay(state.instance) {
		return // 尚未到恢复尝试时间
	}

	// 更新恢复尝试时间
	state.lastAttemptTime = time.Now()
	state.attemptCount++

	// 检查是否可以恢复
	ctx, cancel := context.WithTimeout(context.Background(), m.config.RequestTimeout)
	defer cancel()

	// 尝试恢复
	if m.strategy.CanRecover(ctx, state.instance) {
		if err := m.strategy.Recover(ctx, state.instance); err == nil {
			// 恢复成功，增加成功计数
			state.successCount++
			if state.successCount >= m.config.RecoveryThreshold {
				// 实例已经稳定，可以从恢复列表中移除
				m.mu.Lock()
				delete(m.recoveryStates, state.instance.ID)
				m.mu.Unlock()
			}
		} else {
			// 恢复失败，重置成功计数
			state.successCount = 0
		}
	}
}

// Stop 停止恢复管理器
func (m *RecoveryManager) Stop() {
	close(m.done)
}

// GetRecoveryState 获取实例的恢复状态
func (m *RecoveryManager) GetRecoveryState(instanceID string) (*recoveryState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, ok := m.recoveryStates[instanceID]
	return state, ok
}

// ImmediateRecoveryStrategy 立即恢复策略
// 发现实例故障后立即尝试恢复，适用于短暂故障
type ImmediateRecoveryStrategy struct {
	detector Detector // 用于检测实例是否健康
	config   *Config
}

// NewImmediateRecoveryStrategy 创建立即恢复策略
func NewImmediateRecoveryStrategy(detector Detector, config *Config) *ImmediateRecoveryStrategy {
	return &ImmediateRecoveryStrategy{
		detector: detector,
		config:   config,
	}
}

// CanRecover 判断实例是否可以恢复
func (s *ImmediateRecoveryStrategy) CanRecover(ctx context.Context, instance *naming.Instance) bool {
	// 立即恢复策略总是尝试恢复
	return true
}

// Recover 恢复实例
func (s *ImmediateRecoveryStrategy) Recover(ctx context.Context, instance *naming.Instance) error {
	// 尝试连接实例，检测是否已恢复
	status, err := s.detector.Detect(ctx, instance)
	if err == nil && status == StatusHealthy {
		// 实例已恢复，标记为成功
		return s.detector.MarkSuccess(ctx, instance)
	}
	return err
}

// RecoveryDelay 返回恢复尝试间隔
func (s *ImmediateRecoveryStrategy) RecoveryDelay(instance *naming.Instance) time.Duration {
	return time.Millisecond * 100 // 立即恢复策略使用较短的间隔
}

// GradualRecoveryStrategy 渐进式恢复策略
// 在失败后使用逐渐增加的延迟尝试恢复，避免频繁恢复导致的压力
type GradualRecoveryStrategy struct {
	detector       Detector
	config         *Config
	baseDelay      time.Duration
	maxDelay       time.Duration
	backoffFactor  float64
	attemptCounter map[string]int
	mu             sync.RWMutex
}

// NewGradualRecoveryStrategy 创建渐进式恢复策略
func NewGradualRecoveryStrategy(detector Detector, config *Config) *GradualRecoveryStrategy {
	return &GradualRecoveryStrategy{
		detector:       detector,
		config:         config,
		baseDelay:      time.Second,
		maxDelay:       time.Minute * 5,
		backoffFactor:  1.5,
		attemptCounter: make(map[string]int),
	}
}

// CanRecover 判断实例是否可以恢复
func (s *GradualRecoveryStrategy) CanRecover(ctx context.Context, instance *naming.Instance) bool {
	// 渐进式策略基于上次尝试时间判断
	return true
}

// Recover 恢复实例
func (s *GradualRecoveryStrategy) Recover(ctx context.Context, instance *naming.Instance) error {
	// 尝试连接实例，检测是否已恢复
	status, err := s.detector.Detect(ctx, instance)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err == nil && status == StatusHealthy {
		// 实例已恢复，标记为成功并重置尝试计数
		s.attemptCounter[instance.ID] = 0
		return s.detector.MarkSuccess(ctx, instance)
	}

	// 增加尝试计数
	s.attemptCounter[instance.ID]++
	return err
}

// RecoveryDelay 返回恢复尝试间隔，使用指数退避
func (s *GradualRecoveryStrategy) RecoveryDelay(instance *naming.Instance) time.Duration {
	s.mu.RLock()
	attempts := s.attemptCounter[instance.ID]
	s.mu.RUnlock()

	// 计算基于尝试次数的延迟
	delay := s.baseDelay
	for i := 0; i < attempts; i++ {
		delay = time.Duration(float64(delay) * s.backoffFactor)
		if delay > s.maxDelay {
			delay = s.maxDelay
			break
		}
	}

	return delay
}

// ProbingRecoveryStrategy 探测式恢复策略
// 在恢复时先发送少量请求，确认稳定后再完全恢复
type ProbingRecoveryStrategy struct {
	detector         Detector
	config           *Config
	probingThreshold int               // 探测成功阈值
	probingInterval  time.Duration     // 探测间隔
	probingResults   map[string][]bool // 实例ID -> 探测结果历史
	mu               sync.RWMutex
}

// NewProbingRecoveryStrategy 创建探测式恢复策略
func NewProbingRecoveryStrategy(detector Detector, config *Config) *ProbingRecoveryStrategy {
	return &ProbingRecoveryStrategy{
		detector:         detector,
		config:           config,
		probingThreshold: 3, // 连续3次成功才算恢复
		probingInterval:  time.Second * 5,
		probingResults:   make(map[string][]bool),
	}
}

// CanRecover 判断实例是否可以恢复
func (s *ProbingRecoveryStrategy) CanRecover(ctx context.Context, instance *naming.Instance) bool {
	return true
}

// Recover 恢复实例
func (s *ProbingRecoveryStrategy) Recover(ctx context.Context, instance *naming.Instance) error {
	// 探测实例健康状态
	status, err := s.detector.Detect(ctx, instance)

	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取或初始化探测结果历史
	results, ok := s.probingResults[instance.ID]
	if !ok {
		results = make([]bool, 0, s.probingThreshold)
		s.probingResults[instance.ID] = results
	}

	// 记录探测结果
	success := err == nil && status == StatusHealthy
	s.probingResults[instance.ID] = append(results, success)

	// 保持探测结果历史在阈值范围内
	if len(s.probingResults[instance.ID]) > s.probingThreshold {
		s.probingResults[instance.ID] = s.probingResults[instance.ID][1:]
	}

	// 判断是否达到恢复条件：所有探测都成功
	if len(s.probingResults[instance.ID]) == s.probingThreshold {
		allSuccess := true
		for _, result := range s.probingResults[instance.ID] {
			if !result {
				allSuccess = false
				break
			}
		}

		if allSuccess {
			// 连续探测都成功，认为实例已恢复
			delete(s.probingResults, instance.ID) // 清理探测记录
			return s.detector.MarkSuccess(ctx, instance)
		}
	}

	if success {
		return nil // 当前探测成功但还未达到恢复阈值
	}
	return err // 当前探测失败
}

// RecoveryDelay 返回恢复尝试间隔
func (s *ProbingRecoveryStrategy) RecoveryDelay(instance *naming.Instance) time.Duration {
	return s.probingInterval
}

// NewRecoveryStrategy 创建恢复策略
func NewRecoveryStrategy(config *Config, detector Detector) RecoveryStrategy {
	switch config.RecoveryStrategy {
	case "immediate":
		return NewImmediateRecoveryStrategy(detector, config)
	case "gradual":
		return NewGradualRecoveryStrategy(detector, config)
	case "probing":
		return NewProbingRecoveryStrategy(detector, config)
	default:
		// 默认使用渐进式恢复策略
		return NewGradualRecoveryStrategy(detector, config)
	}
}
