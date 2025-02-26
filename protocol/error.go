package protocol

// 错误定义
var (
	ErrInvalidMagic          = NewError("invalid magic number")
	ErrUnsupportedSerializer = NewError("unsupported serializer type")
)

// Error 自定义错误类型
type Error struct {
	message string
}

func (e *Error) Error() string {
	return e.message
}

// NewError 创建新的错误
func NewError(message string) *Error {
	return &Error{message: message}
}
