package rpc

import (
	"context"
	"net"
	"reflect"

	"github.com/fyerfyer/fyer-rpc/protocol"
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

	methods := make(map[string]reflect.Method)

	// 获取所有方法并验证
	for i := 0; i < serviceValue.Type().NumMethod(); i++ {
		method := serviceValue.Type().Method(i)
		methods[method.Name] = method
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
			return
		}

		// 查找服务
		serviceDesc, ok := s.services[message.Metadata.ServiceName]
		if !ok {
			s.sendError(connection, message.Header.MessageID, "service not found")
			continue
		}

		// 查找方法
		method, ok := serviceDesc.Methods[message.Metadata.MethodName]
		if !ok {
			s.sendError(connection, message.Header.MessageID, "method not found")
			continue
		}

		// 解码参数
		serializer := protocol.GetCodecByType(message.Header.SerializationType)
		if serializer == nil {
			s.sendError(connection, message.Header.MessageID, "unsupported serialization type")
			continue
		}

		reqArg := reflect.New(method.Type.In(2).Elem()).Interface()
		if err := serializer.Decode(message.Payload, reqArg); err != nil {
			s.sendError(connection, message.Header.MessageID, "invalid request parameters")
			continue
		}

		// 创建context
		ctx := context.Background()

		// 调用方法
		results := method.Func.Call([]reflect.Value{
			reflect.ValueOf(serviceDesc.Instance),
			reflect.ValueOf(ctx),
			reflect.ValueOf(reqArg),
		})

		// 处理返回值
		var respPayload []byte
		var respErr string

		if !results[1].IsNil() { // 如果有错误
			respErr = results[1].Interface().(error).Error()
		} else {
			// 序列化返回值
			respPayload, err = serializer.Encode(results[0].Interface())
			if err != nil {
				s.sendError(connection, message.Header.MessageID, "failed to marshal response")
				continue
			}
		}

		// 发送响应
		err = connection.Write(
			message.Metadata.ServiceName,
			message.Metadata.MethodName,
			protocol.TypeResponse,
			message.Header.SerializationType,
			message.Header.MessageID,
			&protocol.Metadata{Error: respErr},
			respPayload,
		)
		if err != nil {
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
