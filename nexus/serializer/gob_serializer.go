package serializer

import (
	"bytes"
	"encoding/gob"
)

type GobSerializer struct {
}

func (gs GobSerializer) Serialize(obj any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf) // 构造编码器，并把数据写进buf中
	if err := encoder.Encode(obj); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (gs GobSerializer) Deserialize(data []byte, obj any) error {
	bufPtr := bytes.NewBuffer(data)   // 返回的类型是 *Buffer，而不是 Buffer。注意一下
	decoder := gob.NewDecoder(bufPtr) // 从 bufPtr 中获取数据
	err := decoder.Decode(obj)
	return err
}
