package protocol

import (
	"encoding/json"
	"testing"

	_ "github.com/fyerfyer/fyer-rpc/protocol/codec" // 导入编解码器
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetadataSerialization(t *testing.T) {
	metadata := &Metadata{
		ServiceName: "TestService",
		MethodName:  "TestMethod",
		Extra:       map[string]string{"key": "value"},
	}

	// JSON序列化
	data, err := json.Marshal(metadata)
	require.NoError(t, err)

	// JSON反序列化
	var decoded Metadata
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// 验证字段
	assert.Equal(t, metadata.ServiceName, decoded.ServiceName)
	assert.Equal(t, metadata.MethodName, decoded.MethodName)
	assert.Equal(t, metadata.Extra["key"], decoded.Extra["key"])
}
