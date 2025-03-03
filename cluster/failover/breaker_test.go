package failover

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/stretchr/testify/assert"
)

// createTestInstance 为测试创建实例
func createTestInstance(id string) *naming.Instance {
	return &naming.Instance{
		ID:      id,
		Service: "test-service",
		Address: "localhost:8080",
		Status:  1,
	}
}

// TestSimpleCircuitBreaker_StateClosed 测试熔断器关闭状态的行为
func TestSimpleCircuitBreaker_StateClosed(t *testing.T) {
	config := NewConfig(
		WithCircuitBreaker(3, time.Second),
	)
	cb := NewSimpleCircuitBreaker(config)
	defer cb.Close()

	instance := createTestInstance("test-1")
	ctx := context.Background()

	// 关闭状态下，所有请求都应该被允许
	for i := 0; i < 5; i++ {
		allow, err := cb.Allow(ctx, instance)
		assert.True(t, allow)
		assert.NoError(t, err)
	}

	// 标记几次失败，但不足以触发熔断
	for i := 0; i < 2; i++ {
		err := cb.MarkFailure(ctx, instance, errors.New("test error"))
		assert.NoError(t, err)
	}

	// 仍然应该允许请求
	allow, err := cb.Allow(ctx, instance)
	assert.True(t, allow)
	assert.NoError(t, err)

	// 再标记一次失败以触发熔断
	err = cb.MarkFailure(ctx, instance, errors.New("test error"))
	assert.NoError(t, err)

	// 现在应该处于开启状态并拒绝请求
	allow, err = cb.Allow(ctx, instance)
	assert.False(t, allow)
	assert.Equal(t, ErrCircuitOpen, err)

	// 检查状态是否开启
	state, err := cb.GetState(instance)
	assert.NoError(t, err)
	assert.Equal(t, StateOpen, state)
}

// TestSimpleCircuitBreaker_StateOpen 测试熔断器开启状态的行为
func TestSimpleCircuitBreaker_StateOpen(t *testing.T) {
	config := NewConfig(
		WithCircuitBreaker(3, 100*time.Millisecond), // 设置较短的超时时间用于测试
	)
	cb := NewSimpleCircuitBreaker(config)
	defer cb.Close()

	instance := createTestInstance("test-1")
	ctx := context.Background()

	// 触发熔断
	for i := 0; i < 3; i++ {
		cb.MarkFailure(ctx, instance, errors.New("test error"))
	}

	// 应该处于开启状态
	allow, err := cb.Allow(ctx, instance)
	assert.False(t, allow)
	assert.Equal(t, ErrCircuitOpen, err)

	// 等待超时时间过期
	time.Sleep(200 * time.Millisecond)

	// 应该转换为半开状态
	state, err := cb.GetState(instance)
	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, state)

	// 半开状态下应该允许有限的请求
	allow, err = cb.Allow(ctx, instance)
	assert.True(t, allow)
	assert.NoError(t, err)
}

// TestSimpleCircuitBreaker_StateHalfOpen 测试熔断器半开状态的行为
func TestSimpleCircuitBreaker_StateHalfOpen(t *testing.T) {
	config := NewConfig(
		WithCircuitBreaker(3, 100*time.Millisecond),
		WithHalfOpenConfig(2, 0.5),
	)
	cb := NewSimpleCircuitBreaker(config)
	defer cb.Close()

	instance := createTestInstance("test-1")
	ctx := context.Background()

	// 触发熔断
	for i := 0; i < 3; i++ {
		cb.MarkFailure(ctx, instance, errors.New("test error"))
	}

	// 等待超时时间
	time.Sleep(200 * time.Millisecond)

	// 此时应该是半开状态
	state, err := cb.GetState(instance)
	assert.NoError(t, err)
	assert.Equal(t, StateHalfOpen, state)

	// 第一个半开状态请求
	allow, err := cb.Allow(ctx, instance)
	assert.True(t, allow)
	assert.NoError(t, err)
	cb.MarkSuccess(ctx, instance)

	// 第二个半开状态请求
	allow, err = cb.Allow(ctx, instance)
	assert.True(t, allow)
	assert.NoError(t, err)
	cb.MarkSuccess(ctx, instance)

	// 此时应该转为关闭状态
	state, err = cb.GetState(instance)
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, state)
}

// TestSimpleCircuitBreaker_HalfOpenFailure 测试半开状态下失败的情况
func TestSimpleCircuitBreaker_HalfOpenFailure(t *testing.T) {
	config := NewConfig(
		WithCircuitBreaker(3, 100*time.Millisecond),
		WithHalfOpenConfig(3, 0.5),
	)
	cb := NewSimpleCircuitBreaker(config)
	defer cb.Close()

	instance := createTestInstance("test-1")
	ctx := context.Background()

	// 触发熔断
	for i := 0; i < 3; i++ {
		cb.MarkFailure(ctx, instance, errors.New("test error"))
	}

	// 等待超时时间
	time.Sleep(200 * time.Millisecond)

	// 第一个请求成功
	allow, _ := cb.Allow(ctx, instance)
	assert.True(t, allow)
	cb.MarkSuccess(ctx, instance)

	// 第二个请求失败
	allow, _ = cb.Allow(ctx, instance)
	assert.True(t, allow)
	cb.MarkFailure(ctx, instance, errors.New("test error"))

	// 应该重新回到开启状态
	state, _ := cb.GetState(instance)
	assert.Equal(t, StateOpen, state)
}

// TestSimpleCircuitBreaker_Reset 测试熔断器重置功能
func TestSimpleCircuitBreaker_Reset(t *testing.T) {
	config := NewConfig(
		WithCircuitBreaker(3, time.Second),
	)
	cb := NewSimpleCircuitBreaker(config)
	defer cb.Close()

	instance := createTestInstance("test-1")
	ctx := context.Background()

	// 触发熔断
	for i := 0; i < 3; i++ {
		cb.MarkFailure(ctx, instance, errors.New("test error"))
	}

	// 验证是否开启
	state, _ := cb.GetState(instance)
	assert.Equal(t, StateOpen, state)

	// 重置熔断器
	err := cb.Reset(instance)
	assert.NoError(t, err)

	// 验证是否关闭
	state, _ = cb.GetState(instance)
	assert.Equal(t, StateClosed, state)
}

// TestSimpleCircuitBreaker_MultipleInstances 测试多实例行为
func TestSimpleCircuitBreaker_MultipleInstances(t *testing.T) {
	config := NewConfig(
		WithCircuitBreaker(3, time.Second),
	)
	cb := NewSimpleCircuitBreaker(config)
	defer cb.Close()

	instance1 := createTestInstance("test-1")
	instance2 := createTestInstance("test-2")
	ctx := context.Background()

	// 触发实例1的熔断
	for i := 0; i < 3; i++ {
		cb.MarkFailure(ctx, instance1, errors.New("test error"))
	}

	// 实例1应该处于开启状态
	allow, _ := cb.Allow(ctx, instance1)
	assert.False(t, allow)

	// 实例2应该仍处于关闭状态
	allow, _ = cb.Allow(ctx, instance2)
	assert.True(t, allow)
}

// TestSimpleCircuitBreaker_Cleanup 测试过期实例的清理
func TestSimpleCircuitBreaker_Cleanup(t *testing.T) {
	// 设置清理间隔和最大空闲时间
	config := NewConfig(
		WithCircuitBreaker(3, time.Second),
	)
	cb := NewSimpleCircuitBreaker(config)
	cb.cleanupInterval = 200 * time.Millisecond
	cb.maxIdleTime = 100 * time.Millisecond
	defer cb.Close()

	instance := createTestInstance("test-1")
	ctx := context.Background()

	// 添加实例到熔断器
	cb.Allow(ctx, instance)

	// 等待清理运行
	time.Sleep(500 * time.Millisecond)

	// 实例应该被清理
	cb.Allow(ctx, instance)
}

// TestCircuitBreakerFactory 测试熔断器工厂函数
func TestCircuitBreakerFactory(t *testing.T) {
	config := NewConfig(
		WithCircuitBreaker(3, time.Second),
	)

	// 测试默认行为
	breaker := NewCircuitBreaker(config)
	assert.NotNil(t, breaker)
}

// TestCircuitBreakerThreadSafety 测试熔断器的线程安全性
func TestCircuitBreakerThreadSafety(t *testing.T) {
	config := NewConfig(
		WithCircuitBreaker(5, time.Second),
	)
	cb := NewSimpleCircuitBreaker(config)
	defer cb.Close()

	instance := createTestInstance("test-1")
	ctx := context.Background()

	// 运行多个goroutine并发执行熔断器操作
	concurrency := 10
	iterations := 100
	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func(routineID int) {
			for j := 0; j < iterations; j++ {
				// 随机选择操作
				operation := j % 3

				switch operation {
				case 0:
					cb.Allow(ctx, instance)
				case 1:
					cb.MarkSuccess(ctx, instance)
				case 2:
					cb.MarkFailure(ctx, instance, errors.New("test error"))
				}
			}
			done <- true
		}(i)
	}

	for i := 0; i < concurrency; i++ {
		<-done
	}
}
