package rpc

import (
	"context"
	"encoding/json"
	"net"
	"reflect"
)

type Server struct {
	services map[string]*ServiceDesc
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]*ServiceDesc),
	}
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
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			return
		}

		// 查找服务
		serviceDesc, ok := s.services[req.ServiceName]
		if !ok {
			encoder.Encode(Response{
				Error: NewRPCError(ErrCodeNotFound, "service not found").Error(),
			})
			continue
		}

		// 查找方法
		method, ok := serviceDesc.Methods[req.MethodName]
		if !ok {
			encoder.Encode(Response{
				Error: NewRPCError(ErrCodeNotFound, "method not found").Error(),
			})
			continue
		}

		// 解码参数
		reqArg := reflect.New(method.Type.In(2).Elem()).Interface()
		if err := json.Unmarshal(req.Args, reqArg); err != nil {
			encoder.Encode(Response{
				Error: NewRPCError(ErrCodeInvalidParam, "invalid request parameters").Error(),
			})
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
		var resp Response
		if !results[1].IsNil() { // 如果有错误
			resp.Error = results[1].Interface().(error).Error()
		} else {
			// 序列化返回值
			data, err := json.Marshal(results[0].Interface())
			if err != nil {
				resp.Error = NewRPCError(ErrCodeInternal, "failed to marshal response").Error()
			} else {
				resp.Data = data
			}
		}

		if err := encoder.Encode(resp); err != nil {
			return
		}
	}
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
