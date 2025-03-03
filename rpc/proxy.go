package rpc

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/fyerfyer/fyer-rpc/cluster/failover"
	"github.com/fyerfyer/fyer-rpc/discovery"
	"github.com/fyerfyer/fyer-rpc/naming"
)

type Proxy struct {
	pool           *ConnPool
	loadBalancer   *discovery.LoadBalancer // 负载均衡器
	enableFailover bool                    // 是否启用故障转移
	failoverConfig *failover.Config        // 故障转移配置
}

// ProxyOption 代理配置选项
type ProxyOption func(*Proxy)

// WithLoadBalancer 设置负载均衡器
func WithLoadBalancer(lb *discovery.LoadBalancer) ProxyOption {
	return func(p *Proxy) {
		p.loadBalancer = lb
	}
}

// WithProxyFailover 设置故障转移
func WithProxyFailover(config *failover.Config) ProxyOption {
	return func(p *Proxy) {
		p.enableFailover = true
		p.failoverConfig = config
	}
}

// InitProxy 初始化服务代理
func InitProxy(address string, target interface{}, opts ...ProxyOption) error {
	// 验证target参数
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return NewRPCError(ErrCodeInvalidParam, "target must be a pointer")
	}
	if targetValue.Elem().Kind() != reflect.Interface {
		return NewRPCError(ErrCodeInvalidParam, "target must be a pointer to interface")
	}

	// 创建连接池
	pool := NewConnPool(address, 10, 5*time.Minute)

	// 创建代理
	proxy := &Proxy{
		pool:           pool,
		enableFailover: false,
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(proxy)
	}

	// 获取目标接口类型
	interfaceType := targetValue.Elem().Type()
	serviceName := interfaceType.Name()

	// 为接口中的每个方法创建代理
	for i := 0; i < interfaceType.NumMethod(); i++ {
		method := interfaceType.Method(i)

		// 创建代理函数
		proxy := reflect.MakeFunc(method.Type, func(args []reflect.Value) []reflect.Value {
			// 从连接池获取连接
			client, err := pool.Get()
			if err != nil {
				return createErrorReturn(method.Type, err)
			}
			defer pool.Put(client)

			// 提取参数
			ctx := args[0].Interface().(context.Context)
			req := args[1].Interface()

			var resp []byte
			var callErr error

			// 根据是否启用了故障转移和负载均衡，选择调用方式
			if proxy.enableFailover && proxy.loadBalancer != nil {
				// 获取当前可用的服务实例
				instances, loadErr := proxy.loadBalancer.GetInstances()
				if loadErr != nil {
					callErr = loadErr
				} else {
					// 通过故障转移调用
					resp, callErr = client.CallWithFailover(ctx, serviceName, method.Name, req, instances)
				}
			} else if proxy.loadBalancer != nil {
				// 使用负载均衡器选择实例
				err := proxy.loadBalancer.SelectWithFailover(ctx, func(ctx context.Context, instance *naming.Instance) error {
					tmpClient, err := NewClient(instance.Address)
					if err != nil {
						return err
					}
					defer tmpClient.Close()

					resp, callErr = tmpClient.CallWithTimeout(ctx, serviceName, method.Name, req)
					return callErr
				})
				if err != nil {
					callErr = err
				}
			} else {
				// 直接调用
				resp, callErr = client.CallWithTimeout(ctx, serviceName, method.Name, req)
			}

			if callErr != nil {
				return createErrorReturn(method.Type, callErr)
			}

			// 解析响应
			result := reflect.New(method.Type.Out(0).Elem()).Interface()
			if err := json.Unmarshal(resp, result); err != nil {
				rpcErr := NewRPCError(ErrCodeInternal, "failed to unmarshal response: "+err.Error())
				return createErrorReturn(method.Type, rpcErr)
			}

			return []reflect.Value{
				reflect.ValueOf(result),
				reflect.Zero(method.Type.Out(1)), // nil error
			}
		})

		// 设置代理方法到接口
		proxyMethod := targetValue.Elem().Method(i)
		proxyMethod.Set(proxy)
	}

	return nil
}

// createErrorReturn 创建错误返回值
func createErrorReturn(methodType reflect.Type, err error) []reflect.Value {
	return []reflect.Value{
		reflect.Zero(methodType.Out(0)),
		reflect.ValueOf(&err).Elem(),
	}
}

// Close 关闭代理的客户端连接
func CloseProxy(target interface{}) error {
	// 暂时不需要特别的清理逻辑，因为连接由池管理
	return nil
}
