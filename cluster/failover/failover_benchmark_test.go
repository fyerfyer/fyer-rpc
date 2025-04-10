package failover

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/naming"
)

var testErr = errors.New("test error")

func BenchmarkFailoverHandler_Execute(b *testing.B) {
	config := &Config{
		MaxRetries:               3,
		RetryInterval:            time.Millisecond * 5,
		MaxRetryDelay:            time.Second * 1,
		RetryBackoff:             1.5,
		RetryJitter:              0.1,
		CircuitBreakThreshold:    5,
		CircuitBreakTimeout:      time.Second * 5,
		HalfOpenMaxCalls:         3,
		HalfOpenSuccessThreshold: 0.6,
		FailureDetectionTime:     time.Second * 5,
		FailureThreshold:         3,
		SuccessThreshold:         2,
		ConnectionTimeout:        time.Millisecond * 100,
		RequestTimeout:           time.Millisecond * 200,
		RecoveryInterval:         time.Second * 5,
		RecoveryTimeout:          time.Second * 30,
		RecoveryStrategy:         "immediate",
		FailoverStrategy:         "next",
	}

	handler, _ := NewFailoverHandler(config)

	detector := newMockDetector()
	handler.detector = detector

	instances := createTestInstances()
	handler.instanceManager = NewInstanceManager(instances)

	successOp := func(ctx context.Context, instance *naming.Instance) error {
		return nil
	}

	failOp := func(ctx context.Context, instance *naming.Instance) error {
		return testErr
	}

	timeoutOp := func(ctx context.Context, instance *naming.Instance) error {
		time.Sleep(time.Millisecond * 10)
		return nil
	}

	b.Run("Success", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			handler.Execute(ctx, instances, successOp)
		}
	})

	b.Run("FailWithRetry", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			handler.Execute(ctx, instances, failOp)
		}
	})

	b.Run("Timeout", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
			handler.Execute(ctx, instances, timeoutOp)
			cancel()
		}
	})
}

func BenchmarkCircuitBreaker_Allow(b *testing.B) {
	config := &Config{
		CircuitBreakThreshold:    5,
		CircuitBreakTimeout:      time.Second * 5,
		HalfOpenMaxCalls:         3,
		HalfOpenSuccessThreshold: 0.6,
	}

	breaker := NewSimpleCircuitBreaker(config)
	instance := createTestInstance("test-1")

	b.Run("Closed", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			breaker.Allow(context.Background(), instance)
		}
	})

	b.Run("Open", func(b *testing.B) {
		ctx := context.Background()
		for i := 0; i < config.CircuitBreakThreshold; i++ {
			breaker.MarkFailure(ctx, instance, errors.New("test error"))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			breaker.Allow(context.Background(), instance)
		}
	})
}

func BenchmarkRetryPolicy(b *testing.B) {
	simplePolicy := NewSimpleRetryPolicy(3, time.Millisecond*10, []string{"test"})
	expPolicy := NewExponentialBackoffRetryPolicy(3, time.Millisecond*10, time.Millisecond*100, 1.5, []string{"test"})
	testErr := errors.New("test error")

	b.Run("SimpleRetryPolicy_ShouldRetry", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			simplePolicy.ShouldRetry(ctx, 1, testErr)
		}
	})

	b.Run("SimpleRetryPolicy_NextBackoff", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			simplePolicy.NextBackoff(1)
		}
	})

	b.Run("ExponentialBackoffRetryPolicy_ShouldRetry", func(b *testing.B) {
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			expPolicy.ShouldRetry(ctx, 1, testErr)
		}
	})

	b.Run("ExponentialBackoffRetryPolicy_NextBackoff", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			expPolicy.NextBackoff(1)
		}
	})
}

func BenchmarkTimeoutDetector_Detect(b *testing.B) {
	config := &Config{
		ConnectionTimeout: time.Millisecond * 100,
		RequestTimeout:    time.Millisecond * 200,
		FailureThreshold:  3,
		SuccessThreshold:  2,
	}

	detector := NewTimeoutDetector(config)
	instance := &naming.Instance{
		ID:      "test-instance",
		Service: "test-service",
		Version: "1.0.0",
		Address: "localhost:12345", // This is likely unreachable
	}

	// Setup a temporary listener to benchmark with a real server
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		b.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()

	// Update the instance with the actual address
	reachableInstance := &naming.Instance{
		ID:      "reachable-instance",
		Service: "test-service",
		Version: "1.0.0",
		Address: listener.Addr().String(),
	}

	// Accept connections in the background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	b.Run("ReachableInstance", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			detector.Detect(context.Background(), reachableInstance)
		}
	})

	b.Run("UnreachableInstance", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			detector.Detect(context.Background(), instance)
		}
	})
}

func BenchmarkInstanceManager_GetInstance(b *testing.B) {
	instances := createTestInstances()
	manager := NewInstanceManager(instances)

	b.Run("Next", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			manager.GetInstance("next")
		}
	})

	b.Run("Random", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			manager.GetInstance("random")
		}
	})

	b.Run("Best", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			manager.GetInstance("best")
		}
	})
}

func BenchmarkParallelFailover(b *testing.B) {
	config := &Config{
		MaxRetries:            3,
		RetryInterval:         time.Millisecond * 5,
		CircuitBreakThreshold: 5,
		CircuitBreakTimeout:   time.Second * 5,
		ConnectionTimeout:     time.Millisecond * 100,
		RequestTimeout:        time.Millisecond * 200,
		FailoverStrategy:      "next",
	}

	handler, _ := NewFailoverHandler(config)

	detector := newMockDetector()
	handler.detector = detector

	instances := createTestInstances()
	handler.instanceManager = NewInstanceManager(instances)

	successOp := func(ctx context.Context, instance *naming.Instance) error {
		return nil
	}

	b.Run("ParallelSuccess", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ctx := context.Background()
				handler.Execute(ctx, instances, successOp)
			}
		})
	})

	failOp := func(ctx context.Context, instance *naming.Instance) error {
		return testErr
	}

	b.Run("ParallelFailure", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				ctx := context.Background()
				handler.Execute(ctx, instances, failOp)
			}
		})
	})
}
