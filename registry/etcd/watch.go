package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/fyerfyer/fyer-rpc/naming"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"sync"
	"sync/atomic"
)

// Watcher 服务监听器
type Watcher struct {
	key       string          // 监听的key
	ctx       context.Context // 上下文
	cancel    func()          // 取消函数
	watchChan clientv3.WatchChan
	client    *clientv3.Client
	eventC    chan []*naming.Instance // 服务实例变更通道
	stopC     chan struct{}           // 停止信号通道
	stopped   int32                   // 是否已停止
	mu        sync.Mutex
}

// newWatcher 创建新的监听器
func newWatcher(client *clientv3.Client, key string) (*Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &Watcher{
		key:       key,
		ctx:       ctx,
		cancel:    cancel,
		client:    client,
		watchChan: client.Watch(ctx, key, clientv3.WithPrefix()),
		eventC:    make(chan []*naming.Instance, 10),
		stopC:     make(chan struct{}),
	}

	// 启动监听协程
	go w.watch()
	return w, nil
}

// watch 监听服务变更
func (w *Watcher) watch() {
	defer close(w.eventC)

	for {
		select {
		case <-w.stopC:
			return
		case resp, ok := <-w.watchChan:
			if !ok {
				return
			}
			if resp.Canceled {
				return
			}

			// 获取最新的服务列表
			instances, err := w.list()
			if err != nil {
				log.Fatalf("list services error: %v", err)
				continue
			}

			// 发送更新的服务列表
			select {
			case w.eventC <- instances:
			case <-w.stopC:
				return
			}
		}
	}
}

// list 获取服务实例列表
func (w *Watcher) list() ([]*naming.Instance, error) {
	resp, err := w.client.Get(w.ctx, w.key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	var instances []*naming.Instance
	for _, kv := range resp.Kvs {
		instance := &naming.Instance{}
		if err := json.Unmarshal(kv.Value, instance); err != nil {
			continue
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

// Next 获取下一次服务更新
func (w *Watcher) Next() ([]*naming.Instance, error) {
	select {
	case instances, ok := <-w.eventC:
		if !ok {
			return nil, errors.New("watcher closed")
		}
		return instances, nil
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	}
}

// Stop 停止监听
func (w *Watcher) Stop() error {
	if !atomic.CompareAndSwapInt32(&w.stopped, 0, 1) {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	close(w.stopC)
	w.cancel()
	return nil
}
