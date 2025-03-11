package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/fyerfyer/fyer-rpc/cluster/failover"
	"github.com/fyerfyer/fyer-rpc/discovery"
	"github.com/fyerfyer/fyer-rpc/naming"
)

type Proxy struct {
	pool            *ConnPool
	loadBalancer    *discovery.LoadBalancer          // 负载均衡器
	enableFailover  bool                             // 是否启用故障转移
	failoverConfig  *failover.Config                 // 故障转移配置
	serviceName     string                           // 服务名称
	failoverHandler *failover.DefaultFailoverHandler // 故障转移处理器
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
		// 创建故障转移处理器
		handler, err := failover.NewFailoverHandler(config)
		if err == nil {
			p.failoverHandler = handler
		}
	}
}

// WithServiceName 设置服务名称
func WithServiceName(serviceName string) ProxyOption {
	return func(p *Proxy) {
		p.serviceName = serviceName
	}
}

// InitProxy 初始化服务代理
func InitProxy(address string, target interface{}, opts ...ProxyOption) error {
	// 验证target参数
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return NewRPCError(ErrCodeInvalidParam, "target must be a pointer")
	}

	// 确保目标是可寻址的
	if !targetValue.Elem().CanSet() {
		return NewRPCError(ErrCodeInvalidParam, "target must be settable")
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

	// 获取目标类型
	targetElem := targetValue.Elem()
	targetType := targetElem.Type()
	serviceName := targetType.Name()
	if proxy.serviceName != "" {
		serviceName = proxy.serviceName
	}

	// 根据目标类型区分处理方式
	if targetElem.Kind() == reflect.Struct {
		// 处理结构体类型
		for i := 0; i < targetType.NumField(); i++ {
			field := targetType.Field(i)

			// 检查字段是否为函数类型
			if field.Type.Kind() == reflect.Func {
				// 创建代理函数
				handleStructField(proxy, targetElem, field, serviceName, pool)
			}
		}
	} else {
		return NewRPCError(ErrCodeInvalidParam, "target must be a pointer to struct")
	}

	return nil
}

// handleStructField 处理结构体字段的代理创建
func handleStructField(proxy *Proxy, targetElem reflect.Value, field reflect.StructField, serviceName string, pool *ConnPool) {
	// 创建代理函数
	proxyFunc := reflect.MakeFunc(field.Type, func(args []reflect.Value) []reflect.Value {
		// 提取参数
		ctx := args[0].Interface().(context.Context)
		req := args[1].Interface()

		// 方法名使用字段名
		methodName := field.Name

		var resp []byte
		var callErr error

		// 使用负载均衡器进行调用
		if proxy.loadBalancer != nil {
			// 获取可用实例列表
			instances, err := proxy.loadBalancer.GetInstances()
			if err != nil {
				return createErrorReturn(field.Type, fmt.Errorf("load balancing failed: %w", err))
			}

			// 使用故障转移机制
			if proxy.enableFailover && proxy.failoverHandler != nil && len(instances) > 0 {
				// 定义调用操作
				operation := func(ctx context.Context, instance *naming.Instance) error {
					// 创建到具体实例的新连接
					client, err := NewClient(instance.Address)
					if err != nil {
						return err
					}
					defer client.Close()

					// 执行RPC调用
					resp, err = client.CallWithTimeout(ctx, serviceName, methodName, req)
					return err
				}

				// 执行带故障转移的调用
				result, err := proxy.failoverHandler.Execute(ctx, instances, operation)
				if err != nil {
					return createErrorReturn(field.Type, fmt.Errorf("failover failed: %w", err))
				}

				// 如果故障转移成功但没有设置响应，重新从返回的实例获取响应
				if result.Success && resp == nil {
					client, err := NewClient(result.Instance.Address)
					if err != nil {
						return createErrorReturn(field.Type, fmt.Errorf("failed to connect to selected instance: %w", err))
					}
					defer client.Close()

					resp, callErr = client.CallWithTimeout(ctx, serviceName, methodName, req)
				} else if !result.Success {
					callErr = fmt.Errorf("no available instances after failover attempts")
				}
			} else {
				// 使用负载均衡但不使用故障转移
				instance, err := proxy.loadBalancer.Select(ctx)
				if err != nil {
					return createErrorReturn(field.Type, fmt.Errorf("load balancing selection failed: %w", err))
				}

				// 创建到选定实例的连接
				client, err := NewClient(instance.Address)
				if err != nil {
					return createErrorReturn(field.Type, fmt.Errorf("failed to connect to selected instance: %w", err))
				}
				defer client.Close()

				// 执行RPC调用
				startTime := time.Now()
				resp, callErr = client.CallWithTimeout(ctx, serviceName, methodName, req)
				duration := time.Since(startTime)

				// 反馈调用结果
				proxy.loadBalancer.Feedback(ctx, instance, duration.Milliseconds(), callErr)
			}
		} else {
			// 从连接池获取连接
			client, err := pool.Get()
			if err != nil {
				return createErrorReturn(field.Type, err)
			}
			defer pool.Put(client)

			// 直接调用原始地址
			resp, callErr = client.CallWithTimeout(ctx, serviceName, methodName, req)
		}

		if callErr != nil {
			return createErrorReturn(field.Type, callErr)
		}

		// 解析响应
		result := reflect.New(field.Type.Out(0).Elem()).Interface()
		if err := json.Unmarshal(resp, result); err != nil {
			return createErrorReturn(field.Type, fmt.Errorf("failed to decode response: %w", err))
		}

		// 返回结果和nil错误
		return []reflect.Value{
			reflect.ValueOf(result),
			reflect.Zero(field.Type.Out(1)),
		}
	})

	// 设置代理函数到结构体字段
	targetElem.FieldByName(field.Name).Set(proxyFunc)
}

// createErrorReturn 创建错误返回值
func createErrorReturn(methodType reflect.Type, err error) []reflect.Value {
	return []reflect.Value{
		reflect.Zero(methodType.Out(0)),
		reflect.ValueOf(&err).Elem(),
	}
}
