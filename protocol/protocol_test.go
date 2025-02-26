package protocol

import (
	"bytes"
	"testing"

	_ "github.com/fyerfyer/fyer-rpc/protocol/codec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProtocolEncodeDecode(t *testing.T) {
	proto := &DefaultProtocol{}

	t.Run("message with metadata", func(t *testing.T) {
		// 创建测试消息
		msg := &Message{
			Header: Header{
				MagicNumber:       MagicNumber,
				Version:           1,
				MessageType:       TypeRequest,
				CompressType:      CompressTypeNone,
				SerializationType: SerializationTypeJSON,
				MessageID:         1,
			},
			Metadata: &Metadata{
				ServiceName: "TestService",
				MethodName:  "TestMethod",
				Extra:       map[string]string{"key": "value"},
			},
			Payload: []byte(`{"test":"data"}`),
		}

		// 编码
		buf := new(bytes.Buffer)
		err := proto.EncodeMessage(msg, buf)
		require.NoError(t, err)

		// 解码
		decoded, err := proto.DecodeMessage(buf)
		require.NoError(t, err)

		// 验证header
		assert.Equal(t, msg.Header.MagicNumber, decoded.Header.MagicNumber)
		assert.Equal(t, msg.Header.MessageID, decoded.Header.MessageID)
		assert.Equal(t, msg.Header.SerializationType, decoded.Header.SerializationType)

		// 验证metadata
		assert.Equal(t, msg.Metadata.ServiceName, decoded.Metadata.ServiceName)
		assert.Equal(t, msg.Metadata.MethodName, decoded.Metadata.MethodName)
		assert.Equal(t, msg.Metadata.Extra["key"], decoded.Metadata.Extra["key"])

		// 验证payload
		assert.Equal(t, msg.Payload, decoded.Payload)
	})
}
