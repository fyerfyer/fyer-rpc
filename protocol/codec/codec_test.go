package codec

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type TestStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestCodec(t *testing.T) {
	testData := &TestStruct{
		Name: "test",
		Age:  20,
	}

	t.Run("JSON codec", func(t *testing.T) {
		// 先检查编解码器是否注册成功
		codec := GetCodec(JSON)
		if codec == nil {
			t.Log("JSON codec is not registered")
			t.FailNow()
		}

		// 编码
		encoded, err := codec.Encode(testData)
		require.NoError(t, err)

		// 解码
		var decoded TestStruct
		err = codec.Decode(encoded, &decoded)
		require.NoError(t, err)

		assert.Equal(t, testData.Name, decoded.Name)
		assert.Equal(t, testData.Age, decoded.Age)
	})

	// 添加错误检查测试
	t.Run("codec not found", func(t *testing.T) {
		codec := GetCodec(Type(33)) // 使用一个未注册的类型
		assert.Nil(t, codec)
	})
}
