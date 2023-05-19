package db

import "encoding/json"

type Codec interface {
	Encode(value interface{}) ([]byte, error)
	Decode(bz []byte, value interface{}) error
}

type JsonCodec struct{}

func (c JsonCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (c JsonCodec) Decode(bz []byte, value interface{}) error {
	return json.Unmarshal(bz, value)
}
