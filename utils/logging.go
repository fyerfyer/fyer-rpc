package utils

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var levelNames = map[LogLevel]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
}

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

// DefaultLogger 默认日志实现
type DefaultLogger struct {
	level  LogLevel
	writer io.Writer
	mu     sync.Mutex
}

// NewLogger 创建新的日志器
func NewLogger(level LogLevel, writer io.Writer) Logger {
	if writer == nil {
		writer = os.Stdout
	}
	return &DefaultLogger{
		level:  level,
		writer: writer,
	}
}

// 默认全局日志器
var (
	defaultLogger     = NewLogger(InfoLevel, os.Stdout)
	defaultLoggerLock sync.RWMutex
)

// SetDefaultLogger 设置默认日志器
func SetDefaultLogger(logger Logger) {
	defaultLoggerLock.Lock()
	defer defaultLoggerLock.Unlock()
	defaultLogger = logger
}

// GetDefaultLogger 获取默认日志器
func GetDefaultLogger() Logger {
	defaultLoggerLock.RLock()
	defer defaultLoggerLock.RUnlock()
	return defaultLogger
}

func (l *DefaultLogger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now().Format("2006-01-02 15:04:05.000")
	levelName := levelNames[level]
	msg := fmt.Sprintf(format, args...)
	logMsg := fmt.Sprintf("[%s] [%s] %s\n", now, levelName, msg)

	_, _ = l.writer.Write([]byte(logMsg))
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug 输出调试级别日志
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	l.log(DebugLevel, format, args...)
}

// Info 输出信息级别日志
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	l.log(InfoLevel, format, args...)
}

// Warn 输出警告级别日志
func (l *DefaultLogger) Warn(format string, args ...interface{}) {
	l.log(WarnLevel, format, args...)
}

// Error 输出错误级别日志
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	l.log(ErrorLevel, format, args...)
}

// Fatal 输出致命级别日志
func (l *DefaultLogger) Fatal(format string, args ...interface{}) {
	l.log(FatalLevel, format, args...)
}

// SetLevel 设置日志级别
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel 获取日志级别
func (l *DefaultLogger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// 全局函数，方便直接使用
func Debug(format string, args ...interface{}) {
	GetDefaultLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetDefaultLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetDefaultLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetDefaultLogger().Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GetDefaultLogger().Fatal(format, args...)
}

// ParseLevel 将字符串转换为日志级别
func ParseLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}
