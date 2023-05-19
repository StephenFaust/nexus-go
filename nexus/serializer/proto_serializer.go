package serializer

import (
	"errors"
	"google.golang.org/protobuf/proto"
)

type ProtoSerializer struct {
}

var NotImplementProtoMessageError = errors.New("param does not implement proto.Message")

func (serializer ProtoSerializer) Serialize(obj any) ([]byte, error) {
	var body proto.Message
	if obj == nil {
		return []byte{}, nil
	}
	var ok bool
	if body, ok = obj.(proto.Message); !ok {
		return nil, NotImplementProtoMessageError
	}
	return proto.Marshal(body)
}

func (serializer ProtoSerializer) Deserialize(data []byte, obj any) error {
	var body proto.Message
	if obj == nil {
		return nil
	}
	var ok bool
	body, ok = obj.(proto.Message)
	if !ok {
		return NotImplementProtoMessageError
	}
	return proto.Unmarshal(data, body)
}
