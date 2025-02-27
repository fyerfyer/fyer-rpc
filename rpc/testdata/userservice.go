package testdata

import (
	"context"
)

// UserService 用户服务结构体
type UserService struct {
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}

// UserServiceImpl 用户服务实现
type UserServiceImpl struct{}

// GetById 获取用户信息
func (s *UserServiceImpl) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	// 这里模拟数据库查询
	if req.Id == 123 {
		return &GetByIdResp{
			User: &User{
				Id:   req.Id,
				Name: "test",
			},
		}, nil
	}
	return &GetByIdResp{}, nil
}
