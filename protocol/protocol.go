package protocol

import (
	"encoding/binary"
	"io"

	"github.com/fyerfyer/fyer-rpc/protocol/codec"
)

// Protocol 协议编解码接口
type Protocol interface {
	EncodeMessage(message *Message, writer io.Writer) error
	DecodeMessage(reader io.Reader) (*Message, error)
}

// DefaultProtocol 默认协议实现
type DefaultProtocol struct{}

// EncodeMessage 编码消息
func (p *DefaultProtocol) EncodeMessage(message *Message, writer io.Writer) error {
	// 写入头部各个字段
	if err := binary.Write(writer, binary.BigEndian, message.Header.MagicNumber); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, message.Header.Version); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, message.Header.MessageType); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, message.Header.CompressType); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, message.Header.SerializationType); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.BigEndian, message.Header.MessageID); err != nil {
		return err
	}

	// 序列化元数据
	var metadataBytes []byte
	var err error
	if message.Metadata != nil {
		codec := GetCodecByType(message.Header.SerializationType)
		if codec == nil {
			return ErrUnsupportedSerializer
		}

		metadataBytes, err = codec.Encode(message.Metadata)
		if err != nil {
			return err
		}
	}

	// 写入元数据长度
	message.Header.MetadataSize = uint32(len(metadataBytes))
	if err := binary.Write(writer, binary.BigEndian, message.Header.MetadataSize); err != nil {
		return err
	}

	// 写入消息体长度
	message.Header.PayloadSize = uint32(len(message.Payload))
	if err := binary.Write(writer, binary.BigEndian, message.Header.PayloadSize); err != nil {
		return err
	}

	// 写入元数据
	if len(metadataBytes) > 0 {
		if _, err := writer.Write(metadataBytes); err != nil {
			return err
		}
	}

	// 写入消息体
	if len(message.Payload) > 0 {
		if _, err := writer.Write(message.Payload); err != nil {
			return err
		}
	}

	return nil
}

// DecodeMessage 解码消息
func (p *DefaultProtocol) DecodeMessage(reader io.Reader) (*Message, error) {
	message := &Message{
		Header: Header{},
	}

	// 读取头部各个字段
	if err := binary.Read(reader, binary.BigEndian, &message.Header.MagicNumber); err != nil {
		return nil, err
	}
	if message.Header.MagicNumber != MagicNumber {
		return nil, ErrInvalidMagic
	}

	if err := binary.Read(reader, binary.BigEndian, &message.Header.Version); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &message.Header.MessageType); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &message.Header.CompressType); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &message.Header.SerializationType); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &message.Header.MessageID); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &message.Header.MetadataSize); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &message.Header.PayloadSize); err != nil {
		return nil, err
	}

	// 读取元数据
	if message.Header.MetadataSize > 0 {
		metadataBytes := make([]byte, message.Header.MetadataSize)
		if _, err := io.ReadFull(reader, metadataBytes); err != nil {
			return nil, err
		}

		codec := GetCodecByType(message.Header.SerializationType)
		if codec == nil {
			return nil, ErrUnsupportedSerializer
		}

		message.Metadata = &Metadata{}
		if err := codec.Decode(metadataBytes, message.Metadata); err != nil {
			return nil, err
		}
	}

	// 读取消息体
	if message.Header.PayloadSize > 0 {
		payload := make([]byte, message.Header.PayloadSize)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return nil, err
		}
		message.Payload = payload
	}

	return message, nil
}

// GetCodecByType 添加工具函数
func GetCodecByType(serializationType uint8) codec.Codec {
	switch serializationType {
	case SerializationTypeJSON:
		return codec.GetCodec(codec.JSON)
	case SerializationTypeProtobuf:
		return codec.GetCodec(codec.Protobuf)
	default:
		return nil
	}
}
