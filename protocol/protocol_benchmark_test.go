package protocol

import (
	"bytes"
	"fmt"
	"testing"

	_ "github.com/fyerfyer/fyer-rpc/protocol/codec"
)

func BenchmarkProtocolEncode(b *testing.B) {
	proto := &DefaultProtocol{}

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

	buf := new(bytes.Buffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		proto.EncodeMessage(msg, buf)
	}
}

func BenchmarkProtocolDecode(b *testing.B) {
	proto := &DefaultProtocol{}

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

	buf := new(bytes.Buffer)
	proto.EncodeMessage(msg, buf)
	encodedData := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(encodedData)
		proto.DecodeMessage(reader)
	}
}

func BenchmarkProtocolEncodeWithDifferentSizes(b *testing.B) {
	sizes := []int{16, 64, 256, 1024, 4096, 16384}

	for _, size := range sizes {
		b.Run(ByteSize(size).String(), func(b *testing.B) {
			proto := &DefaultProtocol{}

			payload := make([]byte, size)
			for i := 0; i < size; i++ {
				payload[i] = byte(i % 256)
			}

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
				Payload: payload,
			}

			buf := new(bytes.Buffer)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				proto.EncodeMessage(msg, buf)
			}
		})
	}
}

func BenchmarkProtocolDecodeWithDifferentSizes(b *testing.B) {
	sizes := []int{16, 64, 256, 1024, 4096, 16384}

	for _, size := range sizes {
		b.Run(ByteSize(size).String(), func(b *testing.B) {
			proto := &DefaultProtocol{}

			payload := make([]byte, size)
			for i := 0; i < size; i++ {
				payload[i] = byte(i % 256)
			}

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
				Payload: payload,
			}

			buf := new(bytes.Buffer)
			proto.EncodeMessage(msg, buf)
			encodedData := buf.Bytes()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				reader := bytes.NewReader(encodedData)
				proto.DecodeMessage(reader)
			}
		})
	}
}

type ByteSize int

func (b ByteSize) String() string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%dB", b)
	case b < 1024*1024:
		return fmt.Sprintf("%dKB", b/1024)
	default:
		return fmt.Sprintf("%dMB", b/(1024*1024))
	}
}

func BenchmarkProtocolWithSerializationTypes(b *testing.B) {
	serializationTypes := map[string]uint8{
		"JSON":     SerializationTypeJSON,
		"Protobuf": SerializationTypeProtobuf,
	}

	for name, serType := range serializationTypes {
		b.Run(name, func(b *testing.B) {
			proto := &DefaultProtocol{}

			msg := &Message{
				Header: Header{
					MagicNumber:       MagicNumber,
					Version:           1,
					MessageType:       TypeRequest,
					CompressType:      CompressTypeNone,
					SerializationType: serType,
					MessageID:         1,
				},
				Metadata: &Metadata{
					ServiceName: "TestService",
					MethodName:  "TestMethod",
					Extra:       map[string]string{"key": "value"},
				},
				Payload: []byte(`{"test":"data"}`),
			}

			buf := new(bytes.Buffer)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf.Reset()
				proto.EncodeMessage(msg, buf)
			}
		})
	}
}

func BenchmarkProtocolRoundTrip(b *testing.B) {
	proto := &DefaultProtocol{}

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		proto.EncodeMessage(msg, buf)
		proto.DecodeMessage(buf)
	}
}
