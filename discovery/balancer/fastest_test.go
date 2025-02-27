package balancer

import (
	"context"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/discovery/metrics"
	"github.com/fyerfyer/fyer-rpc/discovery/mocks"
	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// createMockMetrics 创建模拟的指标收集器
func createMockMetrics(t *testing.T) *mocks.Metrics {
	return mocks.NewMetrics(t)
}

// createTestInstances 创建测试用的服务实例列表
func createTestInstances() []*naming.Instance {
	return []*naming.Instance{
		{
			ID:      "instance-1",
			Service: "test-service",
			Version: "1.0.0",
			Address: "localhost:8001",
			Status:  naming.StatusEnabled,
		},
		{
			ID:      "instance-2",
			Service: "test-service",
			Version: "1.0.0",
			Address: "localhost:8002",
			Status:  naming.StatusEnabled,
		},
		{
			ID:      "instance-3",
			Service: "test-service",
			Version: "1.0.0",
			Address: "localhost:8003",
			Status:  naming.StatusEnabled,
		},
	}
}

func TestFastestBalancer_Initialize(t *testing.T) {
	mockMetrics := createMockMetrics(t)
	instances := createTestInstances()

	// 设置模拟的延迟响应
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8001").
		Return(time.Millisecond*100, nil).Times(1)
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8002").
		Return(time.Millisecond*200, nil).Times(1)
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8003").
		Return(time.Millisecond*300, nil).Times(1)

	// 创建负载均衡器
	balancer := NewFastestBalancer(&Config{
		Type:           FastestResponse,
		MetricsClient:  mockMetrics,
		UpdateInterval: 1,
		RetryTimes:     3,
	})

	// 测试初始化
	err := balancer.Initialize(instances)
	assert.NoError(t, err)
}

func TestFastestBalancer_Select(t *testing.T) {
	mockMetrics := createMockMetrics(t)
	instances := createTestInstances()

	// 设置模拟的延迟响应
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8001").
		Return(time.Millisecond*100, nil).Times(1)
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8002").
		Return(time.Millisecond*200, nil).Times(1)
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8003").
		Return(time.Millisecond*300, nil).Times(1)

	// 创建负载均衡器
	balancer := NewFastestBalancer(&Config{
		Type:           FastestResponse,
		MetricsClient:  mockMetrics,
		UpdateInterval: 1,
		RetryTimes:     3,
	})

	// 初始化
	err := balancer.Initialize(instances)
	assert.NoError(t, err)

	// 测试选择实例
	instance, err := balancer.Select(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "localhost:8001", instance.Address) // 应该选择延迟最低的实例
}

func TestFastestBalancer_Update(t *testing.T) {
	mockMetrics := createMockMetrics(t)
	instances := createTestInstances()

	// 设置初始化时的模拟延迟响应（第一次调用）
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8001").
		Return(time.Millisecond*100, nil).Times(1)
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8002").
		Return(time.Millisecond*200, nil).Times(1)
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8003").
		Return(time.Millisecond*300, nil).Times(1)

	// 设置更新后的模拟延迟响应（第二次调用，只需要两个实例）
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8001").
		Return(time.Millisecond*100, nil).Times(1)
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8002").
		Return(time.Millisecond*200, nil).Times(1)

	// 创建负载均衡器
	balancer := NewFastestBalancer(&Config{
		Type:           FastestResponse,
		MetricsClient:  mockMetrics,
		UpdateInterval: 1,
		RetryTimes:     3,
	})

	// 初始化
	err := balancer.Initialize(instances)
	assert.NoError(t, err)

	// 更新实例列表（移除最后一个实例）
	newInstances := []*naming.Instance{instances[0], instances[1]}
	err = balancer.Update(newInstances)
	assert.NoError(t, err)

	// 验证选择结果
	instance, err := balancer.Select(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "localhost:8001", instance.Address) // 仍应选择延迟最低的实例
}

func TestFastestBalancer_Feedback(t *testing.T) {
	mockMetrics := createMockMetrics(t)
	instances := createTestInstances()

	// 设置模拟延迟响应
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8001").
		Return(time.Millisecond*100, nil).Times(1)
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8002").
		Return(time.Millisecond*200, nil).Times(1)
	mockMetrics.EXPECT().
		GetLatency(mock.Anything, "test-service", "localhost:8003").
		Return(time.Millisecond*300, nil).Times(1)

	// 设置期望的反馈记录
	mockMetrics.EXPECT().
		RecordResponse(mock.Anything, mock.MatchedBy(func(metric *metrics.ResponseMetric) bool {
			return metric.Instance == "localhost:8001" &&
				metric.Status == "success" &&
				metric.Duration == time.Millisecond*50
		})).
		Return(nil).Times(1)

	// 创建负载均衡器
	balancer := NewFastestBalancer(&Config{
		Type:           FastestResponse,
		MetricsClient:  mockMetrics,
		UpdateInterval: 1,
		RetryTimes:     3,
	})

	// 初始化
	err := balancer.Initialize(instances)
	assert.NoError(t, err)

	// 测试反馈
	balancer.Feedback(context.Background(), instances[0], int64(time.Millisecond*50), nil)
}

func TestFastestBalancer_NoAvailableInstances(t *testing.T) {
	mockMetrics := createMockMetrics(t)

	// 创建负载均衡器
	balancer := NewFastestBalancer(&Config{
		Type:           FastestResponse,
		MetricsClient:  mockMetrics,
		UpdateInterval: 1,
		RetryTimes:     3,
	})

	// 初始化空实例列表
	err := balancer.Initialize([]*naming.Instance{})
	assert.NoError(t, err)

	// 测试选择实例
	instance, err := balancer.Select(context.Background())
	assert.Error(t, err)
	assert.Equal(t, ErrNoAvailableInstances, err)
	assert.Nil(t, instance)
}
