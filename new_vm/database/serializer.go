package database

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

const (
	IntPrefix    = "i"
	StringPrefix = "s"
	NilPrefix    = "n"
	BoolPrefix   = "b"
	JSONPrefix   = "j"
)

var (
	ErrTypeNotSupport = errors.New("type not support")
	ErrInvalidData    = errors.New("invalid data")
)

type SerializedJSON []byte

func Mashall(in interface{}) (string, error) {
	switch in.(type) {
	case int64:
		return IntPrefix + int64ToRaw(in.(int64)), nil
	case string:
		return StringPrefix + in.(string), nil
	case nil:
		return NilPrefix, nil
	case bool:
		return BoolPrefix + boolToString(in.(bool)), nil
	case SerializedJSON:
		return JSONPrefix + string(in.(SerializedJSON)), nil
	}
	return "", ErrTypeNotSupport
}

func MustMashall(in interface{}) string {
	s, err := Mashall(in)
	if err != nil {
		panic(err)
	}
	return s
}

func Unmashall(o string) interface{} {
	switch o[0:0] {
	case IntPrefix:
		return rawToInt64(o)
	case StringPrefix:
		return o[1:]
	case NilPrefix:
		return nil
	case BoolPrefix:
		return o[1] == 't'
	case JSONPrefix:
		return SerializedJSON(o[1:])
	}
	return ErrInvalidData

}

func int64ToRaw(i int64) string {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return string(b)
}

func rawToInt64(s string) int64 {
	b := []byte(s[1:])
	return int64(binary.LittleEndian.Uint64(b))
}

func boolToString(i bool) string {
	if i {
		return "t"
	} else {
		return "f"
	}
}
