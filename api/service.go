package api

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/fyerfyer/fyer-rpc/utils"
)

// ServiceInfo 服务信息结构体，用于描述RPC服务
type ServiceInfo struct {
	Name        string            // 服务名称
	Version     string            // 服务版本
	Description string            // 服务描述
	Metadata    map[string]string // 服务元数据
}

// Service RPC服务接口，用户可以选择实现该接口
type Service interface {
	// ServiceInfo 返回服务信息
	ServiceInfo() *ServiceInfo
}

var (
	// ErrServiceNotPointer 服务不是指针类型错误
	ErrServiceNotPointer = errors.New("service must be a pointer to struct")
	// ErrInvalidServiceName 无效的服务名错误
	ErrInvalidServiceName = errors.New("invalid service name")
	// ErrMethodNotFound 方法未找到错误
	ErrMethodNotFound = errors.New("method not found")
	// ErrInvalidMethodSignature 无效的方法签名错误
	ErrInvalidMethodSignature = errors.New("invalid method signature")
)

// ValidateService 验证服务定义是否合法
func ValidateService(service interface{}) error {
	// 检查是否为指针类型
	serviceValue := reflect.ValueOf(service)
	if serviceValue.Kind() != reflect.Ptr {
		return ErrServiceNotPointer
	}

	// 检查方法签名
	if methods, err := utils.GetServiceMethods(service); err != nil || len(methods) == 0 {
		return fmt.Errorf("service has no valid RPC methods: %w", err)
	}

	return nil
}

// ExtractServiceInfo 从服务实例中提取服务信息
func ExtractServiceInfo(service interface{}) (*ServiceInfo, error) {
	// 如果实现了Service接口，直接调用方法
	if s, ok := service.(Service); ok {
		return s.ServiceInfo(), nil
	}

	// 否则通过反射推导服务信息
	serviceValue := reflect.ValueOf(service)
	serviceType := serviceValue.Type().Elem()
	serviceName := serviceType.Name()

	// 移除Impl后缀
	if strings.HasSuffix(serviceName, "Impl") {
		serviceName = serviceName[:len(serviceName)-4]
	}

	return &ServiceInfo{
		Name:     serviceName,
		Version:  "1.0.0", // 默认版本
		Metadata: make(map[string]string),
	}, nil
}
