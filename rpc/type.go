package rpc

import (
	"context"
	"encoding/json"
	"reflect"
)

// Request 表示 RPC 请求
type Request struct {
	ServiceName string          `json:"service_name"`
	MethodName  string          `json:"method_name"`
	Args        json.RawMessage `json:"args"`
}

// Response 表示 RPC 响应
type Response struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error,omitempty"`
}

// ServiceDesc 描述服务的元数据
type ServiceDesc struct {
	ServiceName string
	Methods     map[string]reflect.Method
	Instance    any
}

// MethodType 用于验证方法签名的接口
type MethodType interface {
	// ValidateMethod 验证方法是否符合 (ctx context.Context, req *Request) (*Response, error) 格式
	ValidateMethod(method reflect.Method) bool
}

// DefaultMethodValidator 默认的方法验证器
type DefaultMethodValidator struct{}

func (v *DefaultMethodValidator) ValidateMethod(method reflect.Method) bool {
	// 检查方法参数数量
	if method.Type.NumIn() != 3 { // receiver + ctx + request
		return false
	}
	if method.Type.NumOut() != 2 { // response + error
		return false
	}

	// 检查第一个参数是否为 context.Context
	if !method.Type.In(1).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return false
	}

	// 检查第二个参数是否为指针类型
	if method.Type.In(2).Kind() != reflect.Ptr {
		return false
	}

	// 检查返回值类型
	if method.Type.Out(0).Kind() != reflect.Ptr {
		return false
	}
	if !method.Type.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return false
	}

	return true
}

// InterfaceMethodValidator 接口方法验证器
type InterfaceMethodValidator struct{}

func (v *InterfaceMethodValidator) ValidateMethod(method reflect.Method) bool {
	// 检查方法参数数量
	if method.Type.NumIn() != 2 { // ctx + request
		return false
	}
	if method.Type.NumOut() != 2 { // response + error
		return false
	}

	// 检查第一个参数是否为 context.Context
	if !method.Type.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return false
	}

	// 检查第二个参数是否为指针类型
	if method.Type.In(1).Kind() != reflect.Ptr {
		return false
	}

	// 检查返回值类型
	if method.Type.Out(0).Kind() != reflect.Ptr {
		return false
	}
	if !method.Type.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return false
	}

	return true
}

// RPCError 自定义RPC错误类型
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return e.Message
}

// 预定义错误码
const (
	ErrCodeInternal     = 1000 // 内部错误
	ErrCodeInvalidParam = 1001 // 无效参数
	ErrCodeNotFound     = 1002 // 服务/方法未找到
)

// NewRPCError 创建新的RPC错误
func NewRPCError(code int, message string) *RPCError {
	return &RPCError{
		Code:    code,
		Message: message,
	}
}
