package selector

import (
	"context"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSelector 用于测试的选择器实现
type TestSelector struct {
	name           string
	selectCallback func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
}

func (s *TestSelector) Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
	return s.selectCallback(ctx, instances)
}

func (s *TestSelector) Name() string {
	return s.name
}

func TestChainBasic(t *testing.T) {
	instances := createTestInstances()

	// 创建一个简单的选择器，只返回组A的实例
	groupASelector := &TestSelector{
		name: "group-a",
		selectCallback: func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
			var selected []*naming.Instance
			for _, ins := range instances {
				if ins.Metadata["group"] == "A" {
					selected = append(selected, ins)
				}
			}
			return selected, nil
		},
	}

	// 创建一个选择器，只返回权重大于60的实例
	weightSelector := &TestSelector{
		name: "weight",
		selectCallback: func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
			var selected []*naming.Instance
			for _, ins := range instances {
				if weight := ins.Metadata["weight"]; weight >= "60" {
					selected = append(selected, ins)
				}
			}
			return selected, nil
		},
	}

	// 创建选择器链
	chain := NewChain("test-chain", groupASelector, weightSelector)

	// 测试选择器链
	result, err := chain.Select(context.Background(), instances)
	require.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "instance-1", result[0].ID) // 应该只返回组A中权重大于60的实例
}

func TestChainManagement(t *testing.T) {
	chain := NewChain("test-chain")

	// 测试添加选择器
	selector1 := &TestSelector{name: "test1", selectCallback: func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
		return instances, nil
	}}
	selector2 := &TestSelector{name: "test2", selectCallback: func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
		return instances, nil
	}}

	chain.Add(selector1)
	chain.Add(selector2)
	assert.Equal(t, 2, chain.Len())

	// 测试移除选择器
	removed := chain.Remove("test1")
	assert.True(t, removed)
	assert.Equal(t, 1, chain.Len())

	// 测试重置
	chain.Reset()
	assert.Equal(t, 0, chain.Len())
}

func TestChainEmpty(t *testing.T) {
	chain := NewChain("test-chain")
	instances := createTestInstances()

	// 测试空选择器链
	result, err := chain.Select(context.Background(), instances)
	require.NoError(t, err)
	assert.Equal(t, len(instances), len(result)) // 应该返回所有实例
}

func TestChainContext(t *testing.T) {
	instances := createTestInstances()

	// 创建一个基于context的选择器
	contextSelector := &TestSelector{
		name: "context-based",
		selectCallback: func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
			if group, ok := ctx.Value("group").(string); ok {
				var selected []*naming.Instance
				for _, ins := range instances {
					if ins.Metadata["group"] == group {
						selected = append(selected, ins)
					}
				}
				return selected, nil
			}
			return instances, nil
		},
	}

	chain := NewChain("test-chain", contextSelector)

	// 使用带有group信息的context
	ctx := context.WithValue(context.Background(), "group", "A")
	result, err := chain.Select(ctx, instances)
	require.NoError(t, err)
	assert.Equal(t, 2, len(result)) // 应该返回两个组A的实例

	// 使用普通context
	result, err = chain.Select(context.Background(), instances)
	require.NoError(t, err)
	assert.Equal(t, len(instances), len(result)) // 应该返回所有实例
}

func TestChainConcurrency(t *testing.T) {
	instances := createTestInstances()
	chain := NewChain("test-chain")

	// 添加一个模拟延迟的选择器
	delaySelector := &TestSelector{
		name: "delay",
		selectCallback: func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
			time.Sleep(time.Millisecond)
			return instances, nil
		},
	}
	chain.Add(delaySelector)

	// 并发测试
	concurrency := 10
	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				result, err := chain.Select(context.Background(), instances)
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

func TestChainBuilder(t *testing.T) {
	instances := createTestInstances()

	// 使用构建器模式创建选择器链
	selector1 := &TestSelector{name: "test1", selectCallback: func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
		return instances, nil
	}}
	selector2 := &TestSelector{name: "test2", selectCallback: func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
		return instances, nil
	}}

	chain := NewChainBuilder().
		Add(selector1).
		Add(selector2).
		Build()

	// 验证构建的选择器链
	assert.Equal(t, 2, chain.Len())

	// 测试选择功能
	result, err := chain.Select(context.Background(), instances)
	require.NoError(t, err)
	assert.Equal(t, len(instances), len(result))
}
