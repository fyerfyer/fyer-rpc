package discovery

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
	"github.com/fyerfyer/fyer-rpc/registry/etcd"
)

func BenchmarkDiscovery_GetService(b *testing.B) {
	registry := createBenchmarkTestRegistry(b)
	defer registry.Close()

	discovery := NewDiscovery(registry, time.Second*10)
	defer discovery.Close()

	ctx := context.Background()
	service := "benchmark-service"
	version := "1.0.0"

	instances := []*naming.Instance{
		createTestInstance("bench1", service, version),
		createTestInstance("bench2", service, version),
		createTestInstance("bench3", service, version),
	}

	for _, ins := range instances {
		err := registry.Register(ctx, ins)
		if err != nil {
			b.Fatalf("Failed to register instance: %v", err)
		}
		defer registry.Deregister(ctx, ins)
	}

	time.Sleep(time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := discovery.GetService(ctx, service, version)
		if err != nil {
			b.Fatalf("Failed to get service: %v", err)
		}
	}
}

func BenchmarkDiscovery_GetService_Parallel(b *testing.B) {
	registry := createBenchmarkTestRegistry(b)
	defer registry.Close()

	discovery := NewDiscovery(registry, time.Second*10)
	defer discovery.Close()

	ctx := context.Background()
	service := "benchmark-parallel-service"
	version := "1.0.0"

	instances := []*naming.Instance{
		createTestInstance("bench-p1", service, version),
		createTestInstance("bench-p2", service, version),
		createTestInstance("bench-p3", service, version),
	}

	for _, ins := range instances {
		err := registry.Register(ctx, ins)
		if err != nil {
			b.Fatalf("Failed to register instance: %v", err)
		}
		defer registry.Deregister(ctx, ins)
	}

	time.Sleep(time.Second)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := discovery.GetService(ctx, service, version)
			if err != nil {
				b.Fatalf("Failed to get service: %v", err)
			}
		}
	})
}

func BenchmarkDiscovery_Cache(b *testing.B) {
	registry := createBenchmarkTestRegistry(b)
	defer registry.Close()

	discovery := NewDiscovery(registry, time.Second*10)
	defer discovery.Close()

	ctx := context.Background()
	service := "benchmark-cache-service"
	version := "1.0.0"

	instances := []*naming.Instance{
		createTestInstance("bench-c1", service, version),
		createTestInstance("bench-c2", service, version),
	}

	for _, ins := range instances {
		err := registry.Register(ctx, ins)
		if err != nil {
			b.Fatalf("Failed to register instance: %v", err)
		}
		defer registry.Deregister(ctx, ins)
	}

	time.Sleep(time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		discovery.GetService(ctx, service, version)
	}
}

func BenchmarkDiscovery_MultiService(b *testing.B) {
	registry := createBenchmarkTestRegistry(b)
	defer registry.Close()

	discovery := NewDiscovery(registry, time.Second*10)
	defer discovery.Close()

	ctx := context.Background()
	serviceCount := 10
	instancesPerService := 5

	services := make([]string, serviceCount)
	instancesList := make([][]*naming.Instance, serviceCount)

	for i := 0; i < serviceCount; i++ {
		services[i] = "bench-multi-service-" + string(rune('A'+i))
		instances := make([]*naming.Instance, instancesPerService)

		for j := 0; j < instancesPerService; j++ {
			id := "multi-" + string(rune('A'+i)) + "-" + string(rune('1'+j))
			instances[j] = createTestInstance(id, services[i], "1.0.0")
			err := registry.Register(ctx, instances[j])
			if err != nil {
				b.Fatalf("Failed to register instance: %v", err)
			}
			defer registry.Deregister(ctx, instances[j])
		}

		instancesList[i] = instances
	}

	time.Sleep(time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		serviceIndex := i % serviceCount
		_, err := discovery.GetService(ctx, services[serviceIndex], "1.0.0")
		if err != nil {
			b.Fatalf("Failed to get service: %v", err)
		}
	}
}

func BenchmarkDiscovery_Concurrency(b *testing.B) {
	registry := createBenchmarkTestRegistry(b)
	defer registry.Close()

	discovery := NewDiscovery(registry, time.Second*10)
	defer discovery.Close()

	ctx := context.Background()
	service := "benchmark-concurrency-service"
	version := "1.0.0"

	instanceCount := 10
	instances := make([]*naming.Instance, instanceCount)

	for i := 0; i < instanceCount; i++ {
		instances[i] = createTestInstance("bench-con"+string(rune('1'+i)), service, version)
		err := registry.Register(ctx, instances[i])
		if err != nil {
			b.Fatalf("Failed to register instance: %v", err)
		}
		defer registry.Deregister(ctx, instances[i])
	}

	time.Sleep(time.Second)

	b.ResetTimer()

	concurrency := 50
	var wg sync.WaitGroup
	wg.Add(concurrency)

	operations := b.N
	opPerGoroutine := operations / concurrency

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opPerGoroutine; j++ {
				_, err := discovery.GetService(ctx, service, version)
				if err != nil {
					b.Errorf("Failed to get service: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()
}

func createBenchmarkTestRegistry(b *testing.B) *etcd.EtcdRegistry {
	registry, err := etcd.New(
		etcd.WithEndpoints([]string{"localhost:2379"}),
		etcd.WithDialTimeout(time.Second*5),
		etcd.WithTTL(10),
	)
	if err != nil {
		b.Fatalf("Failed to create registry: %v", err)
	}
	return registry
}
