package main

import (
	"strconv"
	"time"

	"github.com/fyerfyer/fyer-rpc/example/common"
)

// ServerConfigManager 管理服务器的配置
type ServerConfigManager struct {
	// 基本配置
	config *common.ServerConfig
}

// NewServerConfigManager 创建新的服务器配置管理器
func NewServerConfigManager(config *common.ServerConfig) *ServerConfigManager {
	return &ServerConfigManager{
		config: config,
	}
}

// GetConfig 获取当前配置
func (m *ServerConfigManager) GetConfig() *common.ServerConfig {
	return m.config
}

// SetFailureMode 设置故障模式
func (m *ServerConfigManager) SetFailureMode(failAfter int, failDuration time.Duration) {
	m.config.FailAfter = failAfter
	m.config.FailDuration = failDuration
}

// SetRandomFailure 设置随机故障率
func (m *ServerConfigManager) SetRandomFailure(rate float64) {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	m.config.FailRate = rate
}

// CreateClusterConfig 创建集群配置
func CreateClusterConfig(basePort int, count int) []*common.ServerConfig {
	return common.NewClusterServerConfigs(basePort, count)
}

// DefaultClusterConfig 默认集群配置
var DefaultClusterConfig = []*common.ServerConfig{
	{
		ID:           "server-A",
		Address:      "localhost",
		Port:         8001,
		FailAfter:    100, // 处理100个请求后故障
		FailDuration: 10 * time.Second,
	},
	{
		ID:       "server-B",
		Address:  "localhost",
		Port:     8002,
		FailRate: 0.1, // 10%概率随机故障
	},
	{
		ID:      "server-C",
		Address: "localhost",
		Port:    8003,
	},
}

// GetDefaultServerConfig 获取默认服务器配置
func GetDefaultServerConfig() *common.ServerConfig {
	return common.DefaultServerConfig
}

// CloneConfig 克隆配置
func CloneConfig(config *common.ServerConfig) *common.ServerConfig {
	clone := *config
	return &clone
}

// GetServerAddress 获取服务器地址
func GetServerAddress(config *common.ServerConfig) string {
	return config.Address + ":" + strconv.Itoa(config.Port)
}
