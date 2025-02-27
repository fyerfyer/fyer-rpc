package group

import (
	"context"
	"fmt"
	"testing"

	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGroupManager 用于测试的模拟分组管理器
type mockGroupManager struct {
	groups map[string]Group
}

func newMockGroupManager() *mockGroupManager {
	return &mockGroupManager{
		groups: make(map[string]Group),
	}
}

func (m *mockGroupManager) RegisterGroup(group Group) error {
	m.groups[group.Name()] = group
	return nil
}

func (m *mockGroupManager) GetGroup(name string) (Group, error) {
	if group, ok := m.groups[name]; ok {
		return group, nil
	}
	return nil, fmt.Errorf("group not found: %s", name)
}

func (m *mockGroupManager) ListGroups() []Group {
	groups := make([]Group, 0, len(m.groups))
	for _, g := range m.groups {
		groups = append(groups, g)
	}
	return groups
}

func TestRouterBasic(t *testing.T) {
	// 创建测试实例
	instances := []*naming.Instance{
		{
			ID:      "test1",
			Service: "test-service",
			Version: "1.0.0",
			Address: "localhost:8001",
			Metadata: map[string]string{
				"group": "A",
			},
		},
		{
			ID:      "test2",
			Service: "test-service",
			Version: "1.0.0",
			Address: "localhost:8002",
			Metadata: map[string]string{
				"group": "B",
			},
		},
	}

	// 创建分组管理器
	manager := newMockGroupManager()

	// 创建两个测试分组
	groupA, err := NewGroup("A", WithMatcher(&MatchConfig{
		MatchType:  "exact",
		MatchKey:   "group",
		MatchValue: "A",
	}))
	require.NoError(t, err)
	require.NoError(t, manager.RegisterGroup(groupA))

	groupB, err := NewGroup("B", WithMatcher(&MatchConfig{
		MatchType:  "exact",
		MatchKey:   "group",
		MatchValue: "B",
	}))
	require.NoError(t, err)
	require.NoError(t, manager.RegisterGroup(groupB))

	// 创建路由器
	router := NewRouter(manager)

	// 测试路由到分组A
	t.Run("Route to group A", func(t *testing.T) {
		ctx := WithGroup(context.Background(), GroupKey("A"))
		result, err := router.Route(ctx, instances)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "localhost:8001", result[0].Address)
	})

	// 测试路由到分组B
	t.Run("Route to group B", func(t *testing.T) {
		ctx := WithGroup(context.Background(), GroupKey("B"))
		result, err := router.Route(ctx, instances)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "localhost:8002", result[0].Address)
	})

	// 测试不存在的分组
	t.Run("Route to non-existent group", func(t *testing.T) {
		ctx := WithGroup(context.Background(), GroupKey("C"))
		_, err := router.Route(ctx, instances)
		assert.Error(t, err)
	})

	// 测试没有指定分组
	t.Run("Route without group", func(t *testing.T) {
		result, err := router.Route(context.Background(), instances)
		require.NoError(t, err)
		assert.Len(t, result, 2) // 应该返回所有实例
	})
}

func TestRouterChain(t *testing.T) {
	// 创建测试实例
	instances := []*naming.Instance{
		{
			ID:      "test1",
			Service: "test-service",
			Version: "1.0.0",
			Address: "localhost:8001",
			Metadata: map[string]string{
				"group":  "A",
				"weight": "100",
			},
		},
		{
			ID:      "test2",
			Service: "test-service",
			Version: "1.0.0",
			Address: "localhost:8002",
			Metadata: map[string]string{
				"group":  "A",
				"weight": "50",
			},
		},
	}

	// 创建路由链
	chain := NewRouterChain(
		NewTagRouter("group", "A"),
		NewWeightRouter("weight", 80),
	)

	// 测试路由链
	t.Run("Route chain", func(t *testing.T) {
		result, err := chain.Route(context.Background(), instances)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "localhost:8001", result[0].Address)
	})

	// 测试空实例列表
	t.Run("Route with empty instances", func(t *testing.T) {
		_, err := chain.Route(context.Background(), nil)
		assert.Error(t, err)
	})

	// 测试没有匹配的实例
	t.Run("Route with no matching instances", func(t *testing.T) {
		_, err := chain.Route(context.Background(), []*naming.Instance{
			{
				ID:      "test3",
				Service: "test-service",
				Version: "1.0.0",
				Address: "localhost:8003",
				Metadata: map[string]string{
					"group":  "B",
					"weight": "100",
				},
			},
		})
		assert.Error(t, err)
	})
}

func TestSpecializedRouters(t *testing.T) {
	// 测试基于标签的路由器
	t.Run("TagRouter", func(t *testing.T) {
		instances := []*naming.Instance{
			{
				ID:       "test1",
				Metadata: map[string]string{"env": "prod"},
			},
			{
				ID:       "test2",
				Metadata: map[string]string{"env": "test"},
			},
		}

		router := NewTagRouter("env", "prod")
		result, err := router.Route(context.Background(), instances)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "test1", result[0].ID)
	})

	// 测试基于权重的路由器
	t.Run("WeightRouter", func(t *testing.T) {
		instances := []*naming.Instance{
			{
				ID:       "test1",
				Metadata: map[string]string{"weight": "100"}, // 修改权重值大于 80
			},
			{
				ID:       "test2",
				Metadata: map[string]string{"weight": "50"},
			},
		}

		router := NewWeightRouter("weight", 80)
		result, err := router.Route(context.Background(), instances)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "test1", result[0].ID)
	})

	// 测试基于版本的路由器
	t.Run("VersionRouter", func(t *testing.T) {
		instances := []*naming.Instance{
			{
				ID:      "test1",
				Version: "1.0.0",
			},
			{
				ID:      "test2",
				Version: "2.0.0",
			},
		}

		router := NewVersionRouter("1.0.0")
		result, err := router.Route(context.Background(), instances)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "test1", result[0].ID)
	})
}
