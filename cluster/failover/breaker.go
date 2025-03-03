package failover

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
)

// SimpleCircuitBreaker 简单熔断器实现
type SimpleCircuitBreaker struct {
	threshold       int                       // 触发熔断的连续错误阈值
	timeout         time.Duration             // 熔断器打开状态的持续时间
	halfOpenMaxCall int                       // 半开状态最大允许的调用次数
	halfOpenSuccess float64                   // 半开状态转为关闭所需的成功率
	instances       map[string]*instanceState // 实例状态映射
	cleanupInterval time.Duration             // 清理过期实例状态的间隔时间
	maxIdleTime     time.Duration             // 实例状态的最大闲置时间
	mu              sync.RWMutex
	done            chan struct{} // 用于退出清理goroutine
}

// instanceState 记录实例状态
type instanceState struct {
	state            State     // 当前状态
	consecutiveError int32     // 连续错误计数
	lastErrorTime    time.Time // 最后一次错误时间
	openUntil        time.Time // 开路状态持续到的时间点
	halfOpenCounter  int32     // 半开状态计数器
	halfOpenSuccess  int32     // 半开状态成功计数
	halfOpenCalls    int32     // 半开状态调用总数
	lastAccessTime   time.Time // 最后访问时间，用于清理
	openCount        int32     // 熔断器打开次数
	backoffFactor    float64   // 退避因子，用于实现指数退避
	mu               sync.RWMutex
}

// NewSimpleCircuitBreaker 创建简单熔断器
func NewSimpleCircuitBreaker(config *Config) *SimpleCircuitBreaker {
	cb := &SimpleCircuitBreaker{
		threshold:       config.CircuitBreakThreshold,
		timeout:         config.CircuitBreakTimeout,
		halfOpenMaxCall: config.HalfOpenMaxCalls,
		halfOpenSuccess: config.HalfOpenSuccessThreshold,
		instances:       make(map[string]*instanceState),
		cleanupInterval: time.Minute * 10, // 默认10分钟清理一次
		maxIdleTime:     time.Hour,        // 默认1小时未使用则清理
		done:            make(chan struct{}),
	}

	// 启动清理goroutine
	go cb.cleanup()

	return cb
}

// cleanup 定期清理未使用的实例状态
func (b *SimpleCircuitBreaker) cleanup() {
	ticker := time.NewTicker(b.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.done:
			return
		case <-ticker.C:
			b.removeStaleInstances()
		}
	}
}

// removeStaleInstances 移除长时间未使用的实例状态
func (b *SimpleCircuitBreaker) removeStaleInstances() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	for id, state := range b.instances {
		state.mu.RLock()
		lastAccess := state.lastAccessTime
		state.mu.RUnlock()

		// 如果实例状态长时间未被访问且当前是闭合状态，则移除
		if now.Sub(lastAccess) > b.maxIdleTime && state.state == StateClosed {
			delete(b.instances, id)
		}
	}
}

// Close 关闭熔断器，清理资源
func (b *SimpleCircuitBreaker) Close() error {
	close(b.done)
	return nil
}

// Allow 判断调用是否允许通过
func (b *SimpleCircuitBreaker) Allow(ctx context.Context, instance *naming.Instance) (bool, error) {
	state, err := b.getOrCreateState(instance)
	if err != nil {
		return false, err
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	// 更新最后访问时间
	state.lastAccessTime = time.Now()

	// 检查熔断器状态
	switch state.state {
	case StateClosed:
		// 关闭状态，允许所有请求通过
		return true, nil

	case StateOpen:
		// 开路状态，检查是否已经过了超时时间
		if time.Now().After(state.openUntil) {
			// 可以进入半开状态了
			state.mu.RUnlock()
			b.transitionToHalfOpen(instance)
			state.mu.RLock()
			return true, nil
		}
		// 仍在熔断状态，拒绝请求
		return false, ErrCircuitOpen

	case StateHalfOpen:
		// 半开状态，允许有限的请求通过以探测服务是否恢复
		count := atomic.LoadInt32(&state.halfOpenCounter)
		if count < int32(b.halfOpenMaxCall) {
			atomic.AddInt32(&state.halfOpenCounter, 1)
			atomic.AddInt32(&state.halfOpenCalls, 1)
			return true, nil
		}
		return false, ErrCircuitOpen
	}

	return false, ErrCircuitOpen
}

// MarkSuccess 记录成功调用
func (b *SimpleCircuitBreaker) MarkSuccess(ctx context.Context, instance *naming.Instance) error {
	state, err := b.getOrCreateState(instance)
	if err != nil {
		return err
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	// 更新最后访问时间
	state.lastAccessTime = time.Now()

	// 重置连续错误计数
	atomic.StoreInt32(&state.consecutiveError, 0)

	// 处理半开状态下的成功
	if state.state == StateHalfOpen {
		atomic.AddInt32(&state.halfOpenSuccess, 1)

		// 检查是否达到转为关闭状态的条件
		success := atomic.LoadInt32(&state.halfOpenSuccess)
		calls := atomic.LoadInt32(&state.halfOpenCalls)

		if calls > 0 && float64(success)/float64(calls) >= b.halfOpenSuccess {
			// 达到成功率要求，转为关闭状态
			state.state = StateClosed
			atomic.StoreInt32(&state.halfOpenCounter, 0)
			atomic.StoreInt32(&state.halfOpenSuccess, 0)
			atomic.StoreInt32(&state.halfOpenCalls, 0)
			// 重置退避因子
			state.backoffFactor = 1.0
		}
	}

	return nil
}

// MarkFailure 记录失败调用
func (b *SimpleCircuitBreaker) MarkFailure(ctx context.Context, instance *naming.Instance, err error) error {
	state, err := b.getOrCreateState(instance)
	if err != nil {
		return err
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	// 更新最后访问时间和错误时间
	state.lastAccessTime = time.Now()
	state.lastErrorTime = time.Now()

	// 增加连续错误计数
	consecutive := atomic.AddInt32(&state.consecutiveError, 1)

	// 处理不同状态下的失败
	switch state.state {
	case StateClosed:
		// 关闭状态下，检查是否达到熔断阈值
		if int(consecutive) >= b.threshold {
			// 触发熔断，切换到开路状态
			state.state = StateOpen

			// 记录开路次数并应用指数退避
			atomic.AddInt32(&state.openCount, 1)
			if state.backoffFactor < 1.0 {
				state.backoffFactor = 1.0
			}

			// 计算开路时间，使用指数退避增加时长
			timeout := time.Duration(float64(b.timeout) * state.backoffFactor)
			// 限制最大开路时间
			maxTimeout := b.timeout * 10
			if timeout > maxTimeout {
				timeout = maxTimeout
			}

			state.openUntil = time.Now().Add(timeout)
			// 下次开路时间将变长
			state.backoffFactor *= 2.0
		}

	case StateHalfOpen:
		// 半开状态下，任何失败都会使熔断器再次回到开路状态
		state.state = StateOpen

		// 使用退避因子计算下一次开路时长
		timeout := time.Duration(float64(b.timeout) * state.backoffFactor)
		maxTimeout := b.timeout * 10
		if timeout > maxTimeout {
			timeout = maxTimeout
		}

		state.openUntil = time.Now().Add(timeout)
		// 增加退避因子
		state.backoffFactor *= 1.5

		atomic.StoreInt32(&state.halfOpenCounter, 0)
		atomic.StoreInt32(&state.halfOpenSuccess, 0)
		atomic.StoreInt32(&state.halfOpenCalls, 0)
	}

	return nil
}

// GetState 获取熔断器状态
func (b *SimpleCircuitBreaker) GetState(instance *naming.Instance) (State, error) {
	state, err := b.getOrCreateState(instance)
	if err != nil {
		return StateOpen, err
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	// 更新最后访问时间
	state.lastAccessTime = time.Now()

	// 如果是开路状态，检查是否已经超时
	if state.state == StateOpen && time.Now().After(state.openUntil) {
		return StateHalfOpen, nil
	}

	return state.state, nil
}

// Reset 重置熔断器状态
func (b *SimpleCircuitBreaker) Reset(instance *naming.Instance) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 重置该实例的熔断器状态
	b.instances[instance.ID] = &instanceState{
		state:            StateClosed,
		consecutiveError: 0,
		lastErrorTime:    time.Time{},
		openUntil:        time.Time{},
		halfOpenCounter:  0,
		halfOpenSuccess:  0,
		halfOpenCalls:    0,
		lastAccessTime:   time.Now(),
		backoffFactor:    1.0,
	}

	return nil
}

// transitionToHalfOpen 将熔断器状态从开路转为半开
func (b *SimpleCircuitBreaker) transitionToHalfOpen(instance *naming.Instance) {
	state, err := b.getOrCreateState(instance)
	if err != nil {
		return
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.state == StateOpen && time.Now().After(state.openUntil) {
		state.state = StateHalfOpen
		atomic.StoreInt32(&state.halfOpenCounter, 0)
		atomic.StoreInt32(&state.halfOpenSuccess, 0)
		atomic.StoreInt32(&state.halfOpenCalls, 0)
		// 记录状态转换时间
		state.lastAccessTime = time.Now()
	}
}

// getOrCreateState 获取或创建实例状态
func (b *SimpleCircuitBreaker) getOrCreateState(instance *naming.Instance) (*instanceState, error) {
	b.mu.RLock()
	state, ok := b.instances[instance.ID]
	b.mu.RUnlock()

	if ok {
		return state, nil
	}

	// 创建新的实例状态
	b.mu.Lock()
	defer b.mu.Unlock()

	// 再次检查，避免并发创建
	if state, ok = b.instances[instance.ID]; ok {
		return state, nil
	}

	state = &instanceState{
		state:            StateClosed,
		consecutiveError: 0,
		lastErrorTime:    time.Time{},
		openUntil:        time.Time{},
		halfOpenCounter:  0,
		halfOpenSuccess:  0,
		halfOpenCalls:    0,
		lastAccessTime:   time.Now(),
		backoffFactor:    1.0,
	}

	b.instances[instance.ID] = state
	return state, nil
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config *Config) CircuitBreaker {
	// 默认使用简单熔断器
	return NewSimpleCircuitBreaker(config)
}
