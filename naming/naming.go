package naming

import (
	"fmt"
	"strings"
)

// ServiceKey 服务键格式: /fyerrpc/services/{service}/{version}/{instance_id}
const ServiceKey = "/fyerrpc/services/%s/%s/%s"

// Instance 服务实例信息
type Instance struct {
	ID        string            `json:"id"`         // 实例唯一标识
	Service   string            `json:"service"`    // 服务名称
	Version   string            `json:"version"`    // 服务版本
	Address   string            `json:"address"`    // 服务地址
	Metadata  map[string]string `json:"metadata"`   // 元数据
	Status    uint8             `json:"status"`     // 服务状态
	UpdatedAt int64             `json:"updated_at"` // 更新时间
}

const (
	StatusEnabled  = uint8(1) // 服务可用
	StatusDisabled = uint8(0) // 服务不可用
)

// BuildServiceKey 构建服务键
func BuildServiceKey(service, version, instanceID string) string {
	return fmt.Sprintf(ServiceKey, service, version, instanceID)
}

// ParseServiceKey 解析服务键
func ParseServiceKey(key string) (service, version, instanceID string, err error) {
	parts := strings.Split(key, "/")
	if len(parts) != 6 {
		return "", "", "", fmt.Errorf("invalid service key: %s", key)
	}
	return parts[3], parts[4], parts[5], nil
}
