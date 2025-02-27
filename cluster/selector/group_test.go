package selector

import (
	"context"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestInstances() []*naming.Instance {
	return []*naming.Instance{
		{
			ID:      "instance-1",
			Service: "test-service",
			Version: "1.0.0",
			Address: "localhost:8001",
			Metadata: map[string]string{
				"group":  "A",
				"zone":   "zone1",
				"weight": "80",
			},
			Status: naming.StatusEnabled,
		},
		{
			ID:      "instance-2",
			Service: "test-service",
			Version: "1.0.0",
			Address: "localhost:8002",
			Metadata: map[string]string{
				"group":  "B",
				"zone":   "zone1",
				"weight": "50",
			},
			Status: naming.StatusEnabled,
		},
		{
			ID:      "instance-3",
			Service: "test-service",
			Version: "1.0.0",
			Address: "localhost:8003",
			Metadata: map[string]string{
				"group":  "A",
				"zone":   "zone2",
				"weight": "50",
			},
			Status: naming.StatusEnabled,
		},
	}
}

func TestGroupSelector_Select(t *testing.T) {
	instances := createTestInstances()

	tests := []struct {
		name      string
		group     string
		wantCount int
		wantAddrs []string
		wantErr   bool
	}{
		{
			name:      "select group A",
			group:     "A",
			wantCount: 2,
			wantAddrs: []string{"localhost:8001", "localhost:8003"},
			wantErr:   false,
		},
		{
			name:      "select group B",
			group:     "B",
			wantCount: 1,
			wantAddrs: []string{"localhost:8002"},
			wantErr:   false,
		},
		{
			name:      "select non-existent group",
			group:     "C",
			wantCount: 0,
			wantAddrs: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewGroupSelector("test-selector", "group", tt.group)
			result, err := selector.Select(context.Background(), instances)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(result))

			// 验证返回的实例地址
			addrs := make([]string, len(result))
			for i, ins := range result {
				addrs[i] = ins.Address
			}
			assert.ElementsMatch(t, tt.wantAddrs, addrs)
		})
	}
}

func TestABGroupSelector_Select(t *testing.T) {
	instances := createTestInstances()

	tests := []struct {
		name      string
		groupA    string
		groupB    string
		ratio     float64
		forcedCtx context.Context
		wantGroup string
		wantCount int
		wantAddrs []string
		wantErr   bool
	}{
		{
			name:      "select group A with high ratio",
			groupA:    "A",
			groupB:    "B",
			ratio:     0.7,
			wantGroup: "A",
			wantCount: 2,
			wantAddrs: []string{"localhost:8001", "localhost:8003"},
			wantErr:   false,
		},
		{
			name:      "select group B with low ratio",
			groupA:    "A",
			groupB:    "B",
			ratio:     0.3,
			wantGroup: "B",
			wantCount: 1,
			wantAddrs: []string{"localhost:8002"},
			wantErr:   false,
		},
		{
			name:      "forced group selection",
			groupA:    "A",
			groupB:    "B",
			ratio:     0.5,
			forcedCtx: context.WithValue(context.Background(), "ab_group", "A"),
			wantGroup: "A",
			wantCount: 2,
			wantAddrs: []string{"localhost:8001", "localhost:8003"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewABGroupSelector("test-ab-selector", tt.groupA, tt.groupB, tt.ratio)
			ctx := tt.forcedCtx
			if ctx == nil {
				ctx = context.Background()
			}

			result, err := selector.Select(ctx, instances)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(result))

			// 验证返回的实例地址
			addrs := make([]string, len(result))
			for i, ins := range result {
				addrs[i] = ins.Address
				assert.Equal(t, tt.wantGroup, ins.Metadata["group"])
			}
			assert.ElementsMatch(t, tt.wantAddrs, addrs)
		})
	}
}

func TestGroupSelector_EmptyInstances(t *testing.T) {
	selector := NewGroupSelector("test-selector", "group", "A")
	result, err := selector.Select(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestABGroupSelector_EdgeCases(t *testing.T) {
	instances := createTestInstances()

	tests := []struct {
		name    string
		groupA  string
		groupB  string
		ratio   float64
		wantErr bool
	}{
		{
			name:    "invalid ratio below 0",
			groupA:  "A",
			groupB:  "B",
			ratio:   -0.1,
			wantErr: false, // 应该自动修正为0
		},
		{
			name:    "invalid ratio above 1",
			groupA:  "A",
			groupB:  "B",
			ratio:   1.5,
			wantErr: false, // 应该自动修正为1
		},
		{
			name:    "empty group names",
			groupA:  "",
			groupB:  "",
			ratio:   0.5,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := NewABGroupSelector("test-ab-selector", tt.groupA, tt.groupB, tt.ratio)
			_, err := selector.Select(context.Background(), instances)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGroupSelector_Concurrency(t *testing.T) {
	instances := createTestInstances()
	selector := NewGroupSelector("test-selector", "group", "A")

	// 并发测试
	concurrency := 10
	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = selector.Select(context.Background(), instances)
				time.Sleep(time.Millisecond)
			}
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < concurrency; i++ {
		<-done
	}
}
