package utils

import "encoding/json"

func ToJson(v interface{}) (string, error) {
	bz, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bz), nil
}

func MustToJson(v interface{}) string {
	rt, err := ToJson(v)
	if err != nil {
		panic(err)
	}
	return rt
}

func FromJson[T any](jsonString string) (T, error) {
	var v T
	err := json.Unmarshal([]byte(jsonString), &v)
	return v, err
}

func MustFromJson[T any](jsonString string) T {
	var v T
	if err := json.Unmarshal([]byte(jsonString), &v); err != nil {
		panic(err)
	}
	return v
}
