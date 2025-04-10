# `utils` Package

utils包提供了各种辅助功能，如日志、反射和配置等。

### 日志

```go
// Logger 定义日志接口
type Logger interface {
    Debug(format string, args ...interface{})
    Info(format string, args ...interface{})
    Warn(format string, args ...interface{})
    Error(format string, args ...interface{})
    Fatal(format string, args ...interface{})
    SetLevel(level LogLevel)
    GetLevel() LogLevel
}

// NewLogger 创建新的日志器
func NewLogger(level LogLevel, writer io.Writer) Logger
```

全局日志函数：

```go
func Debug(format string, args ...interface{})
func Info(format string, args ...interface{})
func Warn(format string, args ...interface{})
func Error(format string, args ...interface{})
func Fatal(format string, args ...interface{})
```

### 反射工具

```go
// IsExported 判断方法是否导出（公共方法）
func IsExported(name string) bool

// ValidateMethod 验证方法是否符合RPC方法签名
func ValidateMethod(method reflect.Method) error

// GetServiceMethods 获取服务的所有符合RPC方法签名的方法
func GetServiceMethods(service interface{}) (map[string]reflect.Method, error)

// InvokeMethod 调用服务方法
func InvokeMethod(ctx context.Context, instance interface{}, method reflect.Method, arg interface{}) (interface{}, error)
```

### 配置

config子包提供了配置管理功能：

```go
// CommonConfig 通用配置选项
type CommonConfig struct {
    // 日志配置
    LogLevel     utils.LogLevel
    LogOutput    io.Writer
    EnabledDebug bool

    // 协议配置
    SerializationType SerializationType
    CompressType      CompressType
    ProtocolVersion   uint8

    // 超时配置
    DialTimeout    time.Duration
    RequestTimeout time.Duration

    // 其它配置...
}

// ServerConfig 服务器配置
type ServerConfig struct {
    *CommonConfig // 继承通用配置

    // 网络相关配置
    Address         string
    MaxConnections  int
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    ShutdownTimeout time.Duration

    // 其它配置...
}

// ClientConfig 客户端配置
type ClientConfig struct {
    *CommonConfig // 继承通用配置

    // 连接相关配置
    PoolSize        int
    MaxIdle         int
    IdleTimeout     time.Duration
    KeepAlive       bool
    
    // 其它配置...
}
```