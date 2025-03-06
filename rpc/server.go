package rpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"reflect"

	"github.com/fyerfyer/fyer-rpc/protocol"
	"github.com/fyerfyer/fyer-rpc/utils"
)

type Server struct {
	services          map[string]*ServiceDesc
	serializationType uint8
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]*ServiceDesc),
	}
}

// SetSerializationType 设置服务器使用的序列化类型
func (s *Server) SetSerializationType(serializationType uint8) {
	s.serializationType = serializationType
}

// RegisterService 注册服务实例
func (s *Server) RegisterService(service interface{}) error {
	serviceValue := reflect.ValueOf(service)
	if serviceValue.Kind() != reflect.Ptr {
		return NewRPCError(ErrCodeInvalidParam, "service must be a pointer")
	}

	// 获取服务实现类型
	serviceType := serviceValue.Type().Elem()

	// 使用结构体的基础名称（去掉Impl后缀）作为服务名
	serviceName := serviceType.Name()
	if len(serviceName) > 4 && serviceName[len(serviceName)-4:] == "Impl" {
		serviceName = serviceName[:len(serviceName)-4]
	}

	// 使用utils.GetServiceMethods替代手动遍历
	methods, err := utils.GetServiceMethods(service)
	if err != nil {
		return NewRPCError(ErrCodeInvalidParam, fmt.Sprintf("invalid service: %v", err))
	}

	// 如果没有找到有效的方法
	if len(methods) == 0 {
		return NewRPCError(ErrCodeInvalidParam, "service has no valid RPC methods")
	}

	s.services[serviceName] = &ServiceDesc{
		ServiceName: serviceName,
		Methods:     methods,
		Instance:    service,
	}
	return nil
}

// handleConnection 处理每个客户端连接
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	connection := NewConnection(conn)
	defer connection.Close()

	for {
		// 读取请求消息
		message, err := connection.Read()
		if err != nil {
			if err != io.EOF {
				// 只有非EOF错误才记录
				utils.Error("Failed to read message: %v", err)
			}
			return
		}

		// 查找服务
		serviceDesc, ok := s.services[message.Metadata.ServiceName]
		if !ok {
			s.sendError(connection, message.Header.MessageID, fmt.Sprintf("service not found: %s", message.Metadata.ServiceName))
			continue
		}

		// 查找方法
		method, ok := serviceDesc.Methods[message.Metadata.MethodName]
		if !ok {
			s.sendError(connection, message.Header.MessageID, fmt.Sprintf("method not found: %s", message.Metadata.MethodName))
			continue
		}

		// 解码参数
		serializer := protocol.GetCodecByType(message.Header.SerializationType)
		if serializer == nil {
			s.sendError(connection, message.Header.MessageID, "unsupported serialization type")
			continue
		}

		// 创建请求参数实例
		reqType := utils.GetRequestType(method)
		reqArg := reflect.New(reqType).Interface()

		if err := serializer.Decode(message.Payload, reqArg); err != nil {
			s.sendError(connection, message.Header.MessageID, fmt.Sprintf("failed to decode request: %v", err))
			continue
		}

		// 创建context
		ctx := context.Background()

		// 调用方法
		resp, err := utils.InvokeMethod(ctx, serviceDesc.Instance, method, reqArg)
		if err != nil {
			s.sendError(connection, message.Header.MessageID, fmt.Sprintf("method execution error: %v", err))
			continue
		}

		// 序列化响应
		respData, err := serializer.Encode(resp)
		if err != nil {
			s.sendError(connection, message.Header.MessageID, fmt.Sprintf("failed to encode response: %v", err))
			continue
		}

		// 发送响应
		err = connection.Write(
			"",
			"",
			protocol.TypeResponse,
			message.Header.SerializationType,
			message.Header.MessageID,
			nil,
			respData,
		)
		if err != nil {
			utils.Error("Failed to send response: %v", err)
			return
		}
	}
}

func (s *Server) sendError(conn *Connection, messageID uint64, errMsg string) error {
	return conn.Write(
		"",
		"",
		protocol.TypeResponse,
		protocol.SerializationTypeJSON,
		messageID,
		&protocol.Metadata{Error: errMsg},
		nil,
	)
}

func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return NewRPCError(ErrCodeInternal, "failed to start server: "+err.Error())
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go s.handleConnection(conn)
	}
}
