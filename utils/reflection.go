package utils

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
)

var (
	// ErrNotExported 方法未导出错误
	ErrNotExported = errors.New("method is not exported")
	// ErrInvalidMethod 无效方法错误
	ErrInvalidMethod = errors.New("method is invalid")
	// ErrInvalidArgument 无效参数错误
	ErrInvalidArgument = errors.New("argument is invalid")
)

// IsExported 判断方法是否导出（公共方法）
func IsExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// ValidateMethod 验证方法是否符合RPC方法签名
// 符合的签名为: func(ctx context.Context, req *Request) (*Response, error)
func ValidateMethod(method reflect.Method) error {
	// 检查方法名称是否已导出
	if !IsExported(method.Name) {
		return fmt.Errorf("%s: %w", method.Name, ErrNotExported)
	}

	// 检查方法参数数量
	if method.Type.NumIn() != 3 { // receiver + ctx + request
		return fmt.Errorf("%s: %w: expected 3 arguments, got %d", method.Name, ErrInvalidMethod, method.Type.NumIn())
	}

	// 检查返回值数量
	if method.Type.NumOut() != 2 { // response + error
		return fmt.Errorf("%s: %w: expected 2 return values, got %d", method.Name, ErrInvalidMethod, method.Type.NumOut())
	}

	// 检查第一个参数是否为 context.Context
	if !method.Type.In(1).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return fmt.Errorf("%s: %w: first parameter must be context.Context", method.Name, ErrInvalidArgument)
	}

	// 检查第二个参数是否为指针类型
	if method.Type.In(2).Kind() != reflect.Ptr {
		return fmt.Errorf("%s: %w: second parameter must be a pointer", method.Name, ErrInvalidArgument)
	}

	// 检查返回值类型
	if method.Type.Out(0).Kind() != reflect.Ptr {
		return fmt.Errorf("%s: %w: first return value must be a pointer", method.Name, ErrInvalidArgument)
	}
	if !method.Type.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return fmt.Errorf("%s: %w: second return value must be error", method.Name, ErrInvalidArgument)
	}

	return nil
}

// GetServiceMethods 获取服务的所有符合RPC方法签名的方法
func GetServiceMethods(service interface{}) (map[string]reflect.Method, error) {
	serviceValue := reflect.ValueOf(service)
	if serviceValue.Kind() != reflect.Ptr {
		return nil, errors.New("service must be a pointer")
	}

	serviceType := serviceValue.Type()
	methods := make(map[string]reflect.Method)

	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		err := ValidateMethod(method)
		if err != nil {
			// 使用当前包内的Debug函数，而不是外部依赖
			Info("Skipping invalid method %s: %v", method.Name, err)
			continue
		}
		methods[method.Name] = method
	}

	return methods, nil
}

// InvokeMethod 调用服务方法
func InvokeMethod(ctx context.Context, instance interface{}, method reflect.Method, arg interface{}) (interface{}, error) {
	// 创建参数列表
	args := make([]reflect.Value, 3)
	args[0] = reflect.ValueOf(instance) // 服务实例
	args[1] = reflect.ValueOf(ctx)      // 上下文
	args[2] = reflect.ValueOf(arg)      // 请求参数

	// 调用方法
	results := method.Func.Call(args)

	// 处理错误返回值
	if !results[1].IsNil() {
		err := results[1].Interface().(error)
		return nil, err
	}

	// 返回响应
	return results[0].Interface(), nil
}

// GetRequestType 获取方法的请求参数类型
func GetRequestType(method reflect.Method) reflect.Type {
	return method.Type.In(2).Elem()
}

// GetResponseType 获取方法的响应返回值类型
func GetResponseType(method reflect.Method) reflect.Type {
	return method.Type.Out(0).Elem()
}

// CreateInstance 创建指定类型的实例
func CreateInstance(typ reflect.Type) interface{} {
	return reflect.New(typ).Interface()
}
