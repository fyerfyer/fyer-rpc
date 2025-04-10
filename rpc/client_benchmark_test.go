package rpc

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/protocol"
	"github.com/fyerfyer/fyer-rpc/protocol/codec"
)

type benchmarkRequest struct {
	A int
	B int
}

type benchmarkResponse struct {
	Result int
}

type mockServer struct {
	listener net.Listener
	close    chan struct{}
	wg       sync.WaitGroup
}

func setupMockServer(b *testing.B) *mockServer {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("Failed to start mock server: %v", err)
	}

	server := &mockServer{
		listener: listener,
		close:    make(chan struct{}),
	}

	server.wg.Add(1)
	go func() {
		defer server.wg.Done()
		for {
			select {
			case <-server.close:
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					select {
					case <-server.close:
						return
					default:
						b.Logf("Accept error: %v", err)
						time.Sleep(10 * time.Millisecond)
						continue
					}
				}

				server.wg.Add(1)
				go func(c net.Conn) {
					defer server.wg.Done()
					defer c.Close()
					server.handleConnection(b, c)
				}(conn)
			}
		}
	}()

	return server
}

func (s *mockServer) handleConnection(b *testing.B, conn net.Conn) {
	proto := &protocol.DefaultProtocol{}
	jsonCodec := codec.GetCodec(codec.JSON)

	for {
		select {
		case <-s.close:
			return
		default:
			if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				return
			}

			msg, err := proto.DecodeMessage(conn)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue
				}

				return
			}

			if msg.Header.MessageType != protocol.TypeRequest {
				continue
			}

			req := &benchmarkRequest{}
			if err := jsonCodec.Decode(msg.Payload, req); err != nil {
				continue
			}

			resp := &benchmarkResponse{
				Result: req.A + req.B,
			}

			respData, err := jsonCodec.Encode(resp)
			if err != nil {
				continue
			}

			meta := &protocol.Metadata{
				ServiceName: msg.Metadata.ServiceName,
				MethodName:  msg.Metadata.MethodName,
			}

			if err := conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
				return
			}

			if err := proto.EncodeMessage(&protocol.Message{
				Header: protocol.Header{
					MagicNumber:       protocol.MagicNumber,
					Version:           1,
					MessageType:       protocol.TypeResponse,
					CompressType:      protocol.CompressTypeNone,
					SerializationType: protocol.SerializationTypeJSON,
					MessageID:         msg.Header.MessageID,
				},
				Metadata: meta,
				Payload:  respData,
			}, conn); err != nil {
				return
			}
		}
	}
}

func (s *mockServer) stop() {
	close(s.close)
	s.listener.Close()
	s.wg.Wait()
}

func BenchmarkClientCall(b *testing.B) {
	server := setupMockServer(b)
	defer server.stop()

	client, err := NewClient(server.listener.Addr().String())
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	req := &benchmarkRequest{A: 10, B: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := client.Call("TestService", "TestMethod", req)
		if err != nil {
			b.Fatalf("Call failed: %v", err)
		}
		if len(data) == 0 {
			b.Fatal("Response data is empty")
		}
	}
}

func BenchmarkClientParallelCall(b *testing.B) {
	server := setupMockServer(b)
	defer server.stop()

	addr := server.listener.Addr().String()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		client, err := NewClient(addr)
		if err != nil {
			b.Fatalf("Failed to create client: %v", err)
		}
		defer client.Close()

		req := &benchmarkRequest{A: 10, B: 20}
		for pb.Next() {
			data, err := client.Call("TestService", "TestMethod", req)
			if err != nil {
				b.Fatalf("Call failed: %v", err)
			}
			if len(data) == 0 {
				b.Fatal("Response data is empty")
			}
		}
	})
}

func BenchmarkClientPoolCall(b *testing.B) {
	server := setupMockServer(b)
	defer server.stop()

	pool := NewConnPool(server.listener.Addr().String(), 10, time.Minute)
	defer pool.Close()

	req := &benchmarkRequest{A: 10, B: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client, err := pool.Get()
		if err != nil {
			b.Fatalf("Failed to get client from pool: %v", err)
		}

		data, err := client.Call("TestService", "TestMethod", req)
		if err != nil {
			b.Fatalf("Call failed: %v", err)
		}
		if len(data) == 0 {
			b.Fatal("Response data is empty")
		}

		pool.Put(client)
	}
}

func BenchmarkClientPoolParallelCall(b *testing.B) {
	server := setupMockServer(b)
	defer server.stop()

	pool := NewConnPool(server.listener.Addr().String(), 10, time.Minute)
	defer pool.Close()

	req := &benchmarkRequest{A: 10, B: 20}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			client, err := pool.Get()
			if err != nil {
				b.Fatalf("Failed to get client from pool: %v", err)
			}

			data, err := client.Call("TestService", "TestMethod", req)
			if err != nil {
				b.Fatalf("Call failed: %v", err)
			}
			if len(data) == 0 {
				b.Fatal("Response data is empty")
			}

			pool.Put(client)
		}
	})
}

func BenchmarkClientCallWithTimeout(b *testing.B) {
	server := setupMockServer(b)
	defer server.stop()

	client, err := NewClient(server.listener.Addr().String())
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	req := &benchmarkRequest{A: 10, B: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)

		data, err := client.CallWithTimeout(ctx, "TestService", "TestMethod", req)

		cancel()

		if err != nil {
			b.Fatalf("Call failed: %v", err)
		}
		if len(data) == 0 {
			b.Fatal("Response data is empty")
		}
	}
}
