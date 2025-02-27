package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/fyerfyer/fyer-rpc/protocol"
	"github.com/fyerfyer/fyer-rpc/rpc/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCIntegration(t *testing.T) {
	tests := []struct {
		name              string
		serializationType uint8
		addr              string
	}{
		{
			name:              "JSON serialization",
			serializationType: protocol.SerializationTypeJSON,
			addr:              ":8081",
		},
		{
			name:              "Protobuf serialization",
			serializationType: protocol.SerializationTypeProtobuf,
			addr:              ":8082",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 启动服务器
			server := NewServer()
			err := server.RegisterService(&testdata.UserServiceImpl{})
			require.NoError(t, err)

			// 设置服务器使用的序列化类型
			server.SetSerializationType(tt.serializationType)

			go func() {
				err := server.Start(tt.addr)
				if err != nil {
					t.Logf("server stopped: %v", err)
				}
			}()

			time.Sleep(time.Second)

			// 创建客户端代理
			var userService testdata.UserService
			err = InitProxy(":8081", &userService)
			require.NoError(t, err)

			// 测试成功场景
			t.Run("success case", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := userService.GetById(ctx, &testdata.GetByIdReq{Id: 123})
				require.NoError(t, err)
				assert.Equal(t, &testdata.GetByIdResp{
					User: &testdata.User{
						Id:   123,
						Name: "test",
					},
				}, resp)
			})

			// 测试超时场景
			t.Run("timeout case", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond) // 确保超时

				_, err := userService.GetById(ctx, &testdata.GetByIdReq{Id: 123})
				require.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
			})

			// 测试不存在的用户
			t.Run("not found case", func(t *testing.T) {
				ctx := context.Background()
				resp, err := userService.GetById(ctx, &testdata.GetByIdReq{Id: 456})
				require.NoError(t, err)
				assert.Nil(t, resp.User)
			})
		})
	}
}
