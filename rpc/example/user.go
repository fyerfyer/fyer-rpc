package example

// User 用户信息
type User struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

// GetByIdReq 获取用户请求
type GetByIdReq struct {
	Id int64 `json:"id"`
}

// GetByIdResp 获取用户响应
type GetByIdResp struct {
	User *User `json:"user"`
}
