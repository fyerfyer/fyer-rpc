package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createRegistry 创建测试用的registry实例
func createRegistry(t *testing.T) *EtcdRegistry {
	registry, err := New(
		WithEndpoints([]string{"localhost:2379"}),
		WithDialTimeout(time.Second*5),
		WithTTL(10),
	)
	require.NoError(t, err)
	return registry
}

// createTestInstance 创建测试用的服务实例
func createTestInstance(id, service, version, address string) *naming.Instance {
	return &naming.Instance{
		ID:      id,
		Service: service,
		Version: version,
		Address: address,
		Status:  naming.StatusEnabled,
		Metadata: map[string]string{
			"region": "cn-shanghai",
		},
	}
}

func TestRegisterSingle(t *testing.T) {
	registry := createRegistry(t)
	defer registry.Close()

	ctx := context.Background()
	instance := createTestInstance("test-instance-1", "test-service", "1.0.0", "localhost:8080")

	// 测试注册
	err := registry.Register(ctx, instance)
	assert.NoError(t, err)

	// 验证注册结果
	services, err := registry.ListServices(ctx, instance.Service, instance.Version)
	assert.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, instance.ID, services[0].ID)

	// 清理
	err = registry.Deregister(ctx, instance)
	assert.NoError(t, err)
}

func TestRegisterMultiple(t *testing.T) {
	registry := createRegistry(t)
	defer registry.Close()

	ctx := context.Background()
	instances := []*naming.Instance{
		createTestInstance("test-instance-1", "test-service", "1.0.0", "localhost:8080"),
		createTestInstance("test-instance-2", "test-service", "1.0.0", "localhost:8081"),
		createTestInstance("test-instance-3", "test-service", "1.0.0", "localhost:8082"),
	}

	// 注册所有实例
	for _, instance := range instances {
		err := registry.Register(ctx, instance)
		assert.NoError(t, err)
		defer registry.Deregister(ctx, instance)
	}

	time.Sleep(time.Millisecond * 100)

	// 验证注册结果
	services, err := registry.ListServices(ctx, "test-service", "1.0.0")
	assert.NoError(t, err)
	assert.Equal(t, len(instances), len(services))

	// 验证每个实例
	registeredIDs := make(map[string]bool)
	for _, service := range services {
		registeredIDs[service.ID] = true
		assert.Equal(t, "test-service", service.Service)
		assert.Equal(t, "1.0.0", service.Version)
	}

	for _, instance := range instances {
		assert.True(t, registeredIDs[instance.ID])
	}
}

func TestSubscribeSingle(t *testing.T) {
	registry := createRegistry(t)
	defer registry.Close()

	ctx := context.Background()
	instance := createTestInstance("test-subscribe-1", "test-subscribe", "1.0.0", "localhost:8080")

	// 注册实例
	err := registry.Register(ctx, instance)
	require.NoError(t, err)
	defer registry.Deregister(ctx, instance)

	// 订阅服务变更
	watchChan, err := registry.Subscribe(ctx, instance.Service, instance.Version)
	require.NoError(t, err)

	// 验证初始列表
	select {
	case services := <-watchChan:
		assert.Len(t, services, 1)
		assert.Equal(t, instance.ID, services[0].ID)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for initial service list")
	}
}

func TestSubscribeMultiple(t *testing.T) {
	registry := createRegistry(t)
	defer registry.Close()

	ctx := context.Background()
	instance1 := createTestInstance("test-subscribe-1", "test-subscribe", "1.0.0", "localhost:8080")
	instance2 := createTestInstance("test-subscribe-2", "test-subscribe", "1.0.0", "localhost:8081")

	// 注册第一个实例
	err := registry.Register(ctx, instance1)
	require.NoError(t, err)
	defer registry.Deregister(ctx, instance1)

	// 订阅服务变更
	watchChan, err := registry.Subscribe(ctx, instance1.Service, instance1.Version)
	require.NoError(t, err)

	// 验证初始列表
	select {
	case services := <-watchChan:
		assert.Len(t, services, 1)
		assert.Equal(t, instance1.ID, services[0].ID)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for initial service list")
	}

	// 注册第二个实例
	err = registry.Register(ctx, instance2)
	require.NoError(t, err)
	defer registry.Deregister(ctx, instance2)

	// 验证更新后的列表
	select {
	case services := <-watchChan:
		assert.Len(t, services, 2)
		foundIds := make(map[string]bool)
		for _, svc := range services {
			foundIds[svc.ID] = true
		}
		assert.True(t, foundIds[instance1.ID])
		assert.True(t, foundIds[instance2.ID])
	case <-time.After(time.Second * 3):
		t.Fatal("Timeout waiting for service update")
	}
}

func TestHeartbeat(t *testing.T) {
	registry := createRegistry(t)
	defer registry.Close()

	ctx := context.Background()
	instance := createTestInstance("test-heartbeat", "test-service", "1.0.0", "localhost:8082")

	// 注册实例
	err := registry.Register(ctx, instance)
	require.NoError(t, err)
	defer registry.Deregister(ctx, instance)

	// 发送心跳
	err = registry.Heartbeat(ctx, instance)
	assert.NoError(t, err)

	// 等待并验证服务存活
	time.Sleep(time.Second * 5)
	services, err := registry.ListServices(ctx, instance.Service, instance.Version)
	assert.NoError(t, err)

	var found bool
	for _, svc := range services {
		if svc.ID == instance.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "service should exist after heartbeat")
}

func TestDeregister(t *testing.T) {
	registry := createRegistry(t)
	defer registry.Close()

	ctx := context.Background()
	instance := createTestInstance("test-deregister", "test-service", "1.0.0", "localhost:8080")

	// 注册实例
	err := registry.Register(ctx, instance)
	require.NoError(t, err)

	// 注销实例
	err = registry.Deregister(ctx, instance)
	assert.NoError(t, err)

	// 验证注销结果
	services, err := registry.ListServices(ctx, instance.Service, instance.Version)
	assert.NoError(t, err)
	for _, svc := range services {
		assert.NotEqual(t, instance.ID, svc.ID, "service should not exist after deregister")
	}
}
