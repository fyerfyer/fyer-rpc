package rpc

import (
	"context"
	"encoding/json"
	"reflect"
	"time"
)

type Proxy struct {
	pool *ConnPool
}

// InitProxy 初始化服务代理
func InitProxy(address string, target interface{}) error {
	// 验证target参数
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return NewRPCError(ErrCodeInvalidParam, "target must be a pointer")
	}
	if targetValue.Elem().Kind() != reflect.Struct {
		return NewRPCError(ErrCodeInvalidParam, "target must be a pointer to struct")
	}

	// 创建连接池
	pool := NewConnPool(address, 10, 5*time.Minute)

	// 获取目标结构体类型
	structType := targetValue.Elem().Type()
	serviceName := structType.Name()

	// 遍历所有字段
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Type.Kind() != reflect.Func {
			continue
		}

		// 创建代理函数
		proxy := reflect.MakeFunc(field.Type, func(args []reflect.Value) []reflect.Value {
			// 从连接池获取连接
			client, err := pool.Get()
			if err != nil {
				return []reflect.Value{
					reflect.Zero(field.Type.Out(0)),
					reflect.ValueOf(&err).Elem(),
				}
			}
			defer pool.Put(client)

			// 提取参数
			ctx := args[0].Interface().(context.Context)
			req := args[1].Interface()

			// 调用远程方法
			resp, err := client.CallWithTimeout(ctx, serviceName, field.Name, req)
			if err != nil {
				return []reflect.Value{
					reflect.Zero(field.Type.Out(0)),
					reflect.ValueOf(&err).Elem(),
				}
			}

			// 解析响应
			result := reflect.New(field.Type.Out(0).Elem()).Interface()
			if err := json.Unmarshal(resp, result); err != nil {
				rpcErr := NewRPCError(ErrCodeInternal, "failed to unmarshal response: "+err.Error())
				return []reflect.Value{
					reflect.Zero(field.Type.Out(0)),
					reflect.ValueOf(&rpcErr).Elem(),
				}
			}

			return []reflect.Value{
				reflect.ValueOf(result),
				reflect.Zero(field.Type.Out(1)), // nil error
			}
		})

		// 设置代理函数
		targetValue.Elem().Field(i).Set(proxy)
	}

	return nil
}

// Close 关闭代理的客户端连接
func CloseProxy(target interface{}) error {
	// 获取client字段并关闭连接池
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Struct {
		return NewRPCError(ErrCodeInvalidParam, "target must be a pointer to struct")
	}

	// 现在不需要特别的清理逻辑，因为连接由池管理
	return nil
}
