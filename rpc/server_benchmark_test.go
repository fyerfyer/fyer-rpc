package rpc

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/protocol"
)

type BenchService struct{}

type BenchRequest struct {
	Value string
}

type BenchResponse struct {
	Value string
}

func (s *BenchService) Echo(ctx context.Context, req *BenchRequest) (*BenchResponse, error) {
	return &BenchResponse{Value: req.Value}, nil
}

func (s *BenchService) EchoLarge(ctx context.Context, req *BenchRequest) (*BenchResponse, error) {
	return &BenchResponse{Value: req.Value}, nil
}

func startBenchServer(b *testing.B) (*Server, string, func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("Failed to create listener: %v", err)
	}

	addr := listener.Addr().String()
	listener.Close()

	ready := make(chan struct{})

	server := NewServer()
	err = server.RegisterService(&BenchService{})
	if err != nil {
		b.Fatalf("Failed to register service: %v", err)
	}

	errCh := make(chan error, 1)

	go func() {
		close(ready)

		err := server.Start(addr)
		if err != nil {
			errCh <- err
		}
	}()

	<-ready

	time.Sleep(100 * time.Millisecond)

	cleanup := func() {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
		}

		select {
		case err := <-errCh:
			if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
				b.Logf("Server stopped with error: %v", err)
			}
		default:
		}
	}

	return server, addr, cleanup
}

func BenchmarkServer_SingleConnection(b *testing.B) {
	_, addr, cleanup := startBenchServer(b)
	defer cleanup()

	client, err := NewClient(addr)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	req := &BenchRequest{Value: "hello"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Call("BenchService", "Echo", req)
		if err != nil {
			b.Fatalf("Call failed: %v", err)
		}
	}
}

func BenchmarkServer_MultiConnections(b *testing.B) {
	_, addr, cleanup := startBenchServer(b)
	defer cleanup()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		client, err := NewClient(addr)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		req := &BenchRequest{Value: "hello"}
		for pb.Next() {
			_, err := client.Call("BenchService", "Echo", req)
			if err != nil {
				b.Fatalf("Call failed: %v", err)
			}
		}
	})
}

func BenchmarkServer_ConnectionPool(b *testing.B) {
	_, addr, cleanup := startBenchServer(b)
	defer cleanup()

	pool := NewConnPool(addr, 10, 5*time.Minute)
	defer pool.Close()

	req := &BenchRequest{Value: "hello"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client, err := pool.Get()
			if err != nil {
				b.Fatalf("Failed to get client from pool: %v", err)
			}

			_, err = client.Call("BenchService", "Echo", req)
			if err != nil {
				b.Fatalf("Call failed: %v", err)
			}

			pool.Put(client)
		}
	})
}

func BenchmarkServer_LargePayload(b *testing.B) {
	_, addr, cleanup := startBenchServer(b)
	defer cleanup()

	client, err := NewClient(addr)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	largeString := make([]byte, 1024*1024)
	for i := range largeString {
		largeString[i] = byte(i % 256)
	}
	req := &BenchRequest{Value: string(largeString)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Call("BenchService", "EchoLarge", req)
		if err != nil {
			b.Fatalf("Call failed: %v", err)
		}
	}
}

func BenchmarkServer_DifferentSerializationTypes(b *testing.B) {
	serializationTypes := []struct {
		name    string
		typeVal uint8
	}{
		{"JSON", protocol.SerializationTypeJSON},
		{"Protobuf", protocol.SerializationTypeProtobuf},
	}

	for _, st := range serializationTypes {
		b.Run(st.name, func(b *testing.B) {
			server := NewServer()
			server.SetSerializationType(st.typeVal)
			err := server.RegisterService(&BenchService{})
			if err != nil {
				b.Fatalf("Failed to register service: %v", err)
			}

			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				b.Fatalf("Failed to create listener: %v", err)
			}

			addr := listener.Addr().String()
			listener.Close()

			ready := make(chan struct{})

			errCh := make(chan error, 1)

			go func() {
				close(ready)

				err := server.Start(addr)
				if err != nil {
					errCh <- err
				}
			}()

			<-ready

			time.Sleep(100 * time.Millisecond)
			defer func() {
				conn, err := net.Dial("tcp", addr)
				if err == nil {
					conn.Close()
				}

				// Check for any server errors
				select {
				case err := <-errCh:
					if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
						b.Logf("Server stopped with error: %v", err)
					}
				default:
				}
			}()

			client, err := NewClient(addr)
			if err != nil {
				b.Fatalf("Failed to create client: %v", err)
			}
			defer client.Close()

			req := &BenchRequest{Value: "hello"}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Call("BenchService", "Echo", req)
				if err != nil {
					b.Fatalf("Call failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkServer_Concurrent(b *testing.B) {
	_, addr, cleanup := startBenchServer(b)
	defer cleanup()

	numClients := 100
	var clients []*Client
	for i := 0; i < numClients; i++ {
		client, err := NewClient(addr)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
		clients = append(clients, client)
	}
	defer func() {
		for _, client := range clients {
			client.Close()
		}
	}()

	req := &BenchRequest{Value: "hello"}
	var wg sync.WaitGroup
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(numClients)
		for j := 0; j < numClients; j++ {
			go func(idx int) {
				defer wg.Done()
				_, err := clients[idx].Call("BenchService", "Echo", req)
				if err != nil {
					b.Errorf("Call failed: %v", err)
				}
			}(j % numClients)
		}
		wg.Wait()
	}
}
