package rpc

import (
	"sync"
	"time"
)

type ConnPool struct {
	mu          sync.Mutex
	address     string
	maxIdle     int
	idleTimeout time.Duration
	conns       chan *Client
}

func NewConnPool(address string, maxIdle int, idleTimeout time.Duration) *ConnPool {
	return &ConnPool{
		address:     address,
		maxIdle:     maxIdle,
		idleTimeout: idleTimeout,
		conns:       make(chan *Client, maxIdle),
	}
}

func (p *ConnPool) Get() (*Client, error) {
	select {
	case client := <-p.conns:
		if client == nil {
			return p.createConn()
		}
		return client, nil
	default:
		return p.createConn()
	}
}

func (p *ConnPool) Put(client *Client) {
	if client == nil {
		return
	}

	select {
	case p.conns <- client:
		// 成功放回连接池
	default:
		// 连接池满了，直接关闭连接
		client.Close()
	}
}

func (p *ConnPool) createConn() (*Client, error) {
	return NewClient(p.address)
}

func (p *ConnPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.conns)
	for client := range p.conns {
		if client != nil {
			client.Close()
		}
	}
}
