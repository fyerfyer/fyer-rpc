package etcd

import (
	"time"

	"github.com/fyerfyer/fyer-rpc/registry"
)

// Options etcd注册中心特定的配置选项
type Options struct {
	*registry.Options               // 继承通用配置
	Username          string        // etcd用户名
	Password          string        // etcd密码
	CertFile          string        // TLS证书文件
	KeyFile           string        // TLS密钥文件
	TrustedCAFile     string        // TLS CA证书文件
	AutoSyncInterval  time.Duration // 自动同步member列表的间隔时间
	DialKeepAlive     time.Duration // keep-alive探测间隔时间
}

// Option etcd配置选项函数类型
type Option func(*Options)

// DefaultOptions etcd默认配置
var DefaultOptions = &Options{
	Options:          registry.DefaultOptions,
	AutoSyncInterval: time.Minute * 5,
	DialKeepAlive:    time.Second * 30,
}

// WithUsername 设置etcd用户名
func WithUsername(username string) Option {
	return func(o *Options) {
		o.Username = username
	}
}

// WithPassword 设置etcd密码
func WithPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// WithTLSConfig 设置TLS配置
func WithTLSConfig(certFile, keyFile, caFile string) Option {
	return func(o *Options) {
		o.CertFile = certFile
		o.KeyFile = keyFile
		o.TrustedCAFile = caFile
	}
}

// WithAutoSyncInterval 设置自动同步间隔
func WithAutoSyncInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.AutoSyncInterval = interval
	}
}

// WithDialKeepAlive 设置keep-alive间隔
func WithDialKeepAlive(keepalive time.Duration) Option {
	return func(o *Options) {
		o.DialKeepAlive = keepalive
	}
}

// WithEndpoints 设置etcd endpoints
func WithEndpoints(endpoints []string) Option {
	return func(o *Options) {
		o.Options.Endpoints = endpoints // 设置通用配置
	}
}

// WithDialTimeout 设置连接超时时间
func WithDialTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Options.DialTimeout = timeout // 设置通用配置
	}
}

// WithTTL 设置服务租约时间
func WithTTL(ttl int64) Option {
	return func(o *Options) {
		o.Options.TTL = ttl // 设置通用配置
	}
}
