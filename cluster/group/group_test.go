package group

import (
	"context"
	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func createTestInstance(id string, metadata map[string]string) *naming.Instance {
	return &naming.Instance{
		ID:       id,
		Service:  "test-service",
		Version:  "1.0.0",
		Address:  "localhost:8080",
		Status:   naming.StatusEnabled,
		Metadata: metadata,
	}
}

func TestNewGroup(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		opts      []Option
		wantErr   bool
	}{
		{
			name:      "valid group",
			groupName: "test-group",
			opts: []Option{
				WithType("test"),
				WithWeight(80),
			},
			wantErr: false,
		},
		{
			name:      "empty group name",
			groupName: "",
			opts:      nil,
			wantErr:   true,
		},
		{
			name:      "invalid matcher",
			groupName: "test-group",
			opts: []Option{
				WithMatcher(&MatchConfig{
					MatchType:  "unknown",
					MatchKey:   "key",
					MatchValue: "value",
				}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := NewGroup(tt.groupName, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.groupName, group.Name())
		})
	}
}

func TestGroupMatch(t *testing.T) {
	tests := []struct {
		name     string
		matcher  *MatchConfig
		instance *naming.Instance
		want     bool
	}{
		{
			name: "exact match success",
			matcher: &MatchConfig{
				MatchType:  "exact",
				MatchKey:   "version",
				MatchValue: "v1",
			},
			instance: createTestInstance("test1", map[string]string{
				"version": "v1",
			}),
			want: true,
		},
		{
			name: "exact match failure",
			matcher: &MatchConfig{
				MatchType:  "exact",
				MatchKey:   "version",
				MatchValue: "v1",
			},
			instance: createTestInstance("test2", map[string]string{
				"version": "v2",
			}),
			want: false,
		},
		{
			name: "prefix match success",
			matcher: &MatchConfig{
				MatchType:  "prefix",
				MatchKey:   "env",
				MatchValue: "prod",
			},
			instance: createTestInstance("test3", map[string]string{
				"env": "prod-east",
			}),
			want: true,
		},
		{
			name: "regex match success",
			matcher: &MatchConfig{
				MatchType:  "regex",
				MatchKey:   "region",
				MatchValue: "^us-.*$",
			},
			instance: createTestInstance("test4", map[string]string{
				"region": "us-east-1",
			}),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := NewGroup("test-group", WithMatcher(tt.matcher))
			require.NoError(t, err)
			assert.Equal(t, tt.want, group.Match(tt.instance))
		})
	}
}

func TestGroupSelect(t *testing.T) {
	instances := []*naming.Instance{
		createTestInstance("test1", map[string]string{"env": "prod"}),
		createTestInstance("test2", map[string]string{"env": "staging"}),
		createTestInstance("test3", map[string]string{"env": "prod"}),
	}

	tests := []struct {
		name      string
		matcher   *MatchConfig
		wantCount int
		wantErr   bool
	}{
		{
			name: "select prod instances",
			matcher: &MatchConfig{
				MatchType:  "exact",
				MatchKey:   "env",
				MatchValue: "prod",
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "select staging instances",
			matcher: &MatchConfig{
				MatchType:  "exact",
				MatchKey:   "env",
				MatchValue: "staging",
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "select non-existent instances",
			matcher: &MatchConfig{
				MatchType:  "exact",
				MatchKey:   "env",
				MatchValue: "test",
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, err := NewGroup("test-group", WithMatcher(tt.matcher))
			require.NoError(t, err)

			selected, err := group.Select(context.Background(), instances)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, len(selected))
		})
	}
}

func TestGroupUpdateConfig(t *testing.T) {
	group, err := NewGroup("test-group",
		WithType("test"),
		WithWeight(80),
		WithMatcher(&MatchConfig{
			MatchType:  "exact",
			MatchKey:   "version",
			MatchValue: "v1",
		}),
	)
	require.NoError(t, err)

	// Test initial configuration
	assert.Equal(t, 80, group.GetWeight())
	instance := createTestInstance("test1", map[string]string{"version": "v1"})
	assert.True(t, group.Match(instance))

	// Update configuration
	err = group.UpdateConfig(
		WithWeight(90),
		WithMatcher(&MatchConfig{
			MatchType:  "exact",
			MatchKey:   "version",
			MatchValue: "v2",
		}),
	)
	require.NoError(t, err)

	// Test updated configuration
	assert.Equal(t, 90, group.GetWeight())
	assert.False(t, group.Match(instance))
	instance = createTestInstance("test2", map[string]string{"version": "v2"})
	assert.True(t, group.Match(instance))
}

func TestGroupConcurrency(t *testing.T) {
	group, err := NewGroup("test-group")
	require.NoError(t, err)

	instances := []*naming.Instance{
		createTestInstance("test1", map[string]string{"env": "prod"}),
		createTestInstance("test2", map[string]string{"env": "staging"}),
	}

	// Test concurrent access
	concurrency := 10
	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = group.Select(context.Background(), instances)
				_ = group.Match(instances[0])
				_ = group.UpdateConfig(WithWeight(j % 100))
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}
}
