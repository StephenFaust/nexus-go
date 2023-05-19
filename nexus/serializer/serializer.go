package serializer

type Serializer interface {
	Serialize(obj any) ([]byte, error)
	Deserialize(data []byte, obj any) error
}
