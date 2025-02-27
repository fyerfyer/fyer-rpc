package discovery

import (
	"context"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/fyerfyer/fyer-rpc/registry/etcd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRegistry 创建用于测试的注册中心
func createTestRegistry(t *testing.T) *etcd.EtcdRegistry {
	registry, err := etcd.New(
		etcd.WithEndpoints([]string{"localhost:2379"}),
		etcd.WithDialTimeout(time.Second*5),
		etcd.WithTTL(10),
	)
	require.NoError(t, err)
	return registry
}

// createTestInstance 创建测试服务实例
func createTestInstance(id string, service string, version string) *naming.Instance {
	return &naming.Instance{
		ID:      id,
		Service: service,
		Version: version,
		Address: "localhost:808" + id[len(id)-1:],
		Status:  naming.StatusEnabled,
		Metadata: map[string]string{
			"region": "test-region",
		},
	}
}

func TestDiscoveryGetService(t *testing.T) {
	registry := createTestRegistry(t)
	defer registry.Close()

	// 创建服务发现实例
	discovery := NewDiscovery(registry, time.Second*10)
	defer discovery.Close()

	ctx := context.Background()
	service := "test-service"
	version := "1.0.0"

	// 注册测试实例
	instances := []*naming.Instance{
		createTestInstance("test1", service, version),
		createTestInstance("test2", service, version),
		createTestInstance("test3", service, version),
	}

	// 注册服务实例
	for _, ins := range instances {
		err := registry.Register(ctx, ins)
		require.NoError(t, err)
		defer registry.Deregister(ctx, ins)
	}

	// 等待服务注册完成
	time.Sleep(time.Second)

	// 测试获取服务
	t.Run("GetService", func(t *testing.T) {
		services, err := discovery.GetService(ctx, service, version)
		assert.NoError(t, err)
		assert.Equal(t, len(instances), len(services))

		// 验证返回的实例信息
		foundInstances := make(map[string]bool)
		for _, svc := range services {
			foundInstances[svc.ID] = true
			assert.Equal(t, service, svc.Service)
			assert.Equal(t, version, svc.Version)
			assert.Equal(t, naming.StatusEnabled, svc.Status)
		}

		// 确保所有注册的实例都被找到
		for _, ins := range instances {
			assert.True(t, foundInstances[ins.ID])
		}
	})
}

func TestDiscoveryWatch(t *testing.T) {
	registry := createTestRegistry(t)
	defer registry.Close()

	discovery := NewDiscovery(registry, time.Second*10)
	defer discovery.Close()

	ctx := context.Background()
	service := "test-watch-service"
	version := "1.0.0"

	// 创建初始实例
	instance1 := createTestInstance("watch1", service, version)

	// 注册第一个实例
	err := registry.Register(ctx, instance1)
	require.NoError(t, err)
	defer registry.Deregister(ctx, instance1)

	// 创建监听器
	watcher, err := discovery.Watch(ctx, service, version)
	require.NoError(t, err)
	defer watcher.Stop() // 确保在测试结束时停止 watcher

	// 测试初始服务列表
	t.Run("InitialWatch", func(t *testing.T) {
		services, err := watcher.Next()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(services))
		assert.Equal(t, instance1.ID, services[0].ID)
	})

	// 测试服务更新
	t.Run("WatchUpdate", func(t *testing.T) {
		// 注册新实例
		instance2 := createTestInstance("watch2", service, version)
		err := registry.Register(ctx, instance2)
		require.NoError(t, err)
		defer registry.Deregister(ctx, instance2)

		// 等待服务更新
		time.Sleep(time.Second)

		// 验证更新后的服务列表
		services, err := watcher.Next()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(services))

		// 验证实例ID
		ids := make(map[string]bool)
		for _, svc := range services {
			ids[svc.ID] = true
		}
		assert.True(t, ids[instance1.ID])
		assert.True(t, ids[instance2.ID])
	})

	// 在测试结束前等待一下，确保所有操作完成
	time.Sleep(time.Millisecond * 100)
}

func TestDiscoveryCache(t *testing.T) {
	registry := createTestRegistry(t)
	discovery := NewDiscovery(registry, time.Second*2)

	ctx := context.Background()
	service := "test-cache-service"
	version := "1.0.0"
	instance := createTestInstance("cache1", service, version)

	// 注册服务实例
	err := registry.Register(ctx, instance)
	require.NoError(t, err)

	// 使用 cleanup 函数确保按正确的顺序清理资源
	t.Cleanup(func() {
		// 先注销服务实例
		_ = registry.Deregister(ctx, instance)
		// 等待一小段时间确保注销操作完成
		time.Sleep(time.Millisecond * 100)
		// 然后关闭 discovery
		_ = discovery.Close()
		// 最后关闭 registry
		_ = registry.Close()
	})

	t.Run("CacheHit", func(t *testing.T) {
		// 第一次获取服务（将填充缓存）
		services1, err := discovery.GetService(ctx, service, version)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(services1))

		// 立即再次获取（应该命中缓存）
		services2, err := discovery.GetService(ctx, service, version)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(services2))
		assert.Equal(t, services1[0].ID, services2[0].ID)
	})

	t.Run("CacheExpire", func(t *testing.T) {
		// 等待缓存过期
		time.Sleep(time.Second * 3)

		// 获取服务（应该重新从注册中心获取）
		services, err := discovery.GetService(ctx, service, version)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(services))
		assert.Equal(t, instance.ID, services[0].ID)
	})
}

func TestDiscoveryClose(t *testing.T) {
	registry := createTestRegistry(t)
	defer registry.Close()

	discovery := NewDiscovery(registry, time.Second*10)

	// 测试正常关闭
	t.Run("NormalClose", func(t *testing.T) {
		err := discovery.Close()
		assert.NoError(t, err)
	})

	// 测试关闭后的操作
	t.Run("OperationAfterClose", func(t *testing.T) {
		ctx := context.Background()
		_, err := discovery.GetService(ctx, "test-service", "1.0.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "discovery is closed")
	})

	// 测试重复关闭
	t.Run("DoubleClose", func(t *testing.T) {
		err := discovery.Close()
		assert.NoError(t, err)
	})
}
