package database

import (
	"encoding/binary"
	"errors"
	"strings"
)

// const prefixs
const (
	IntPrefix       = "i"
	StringPrefix    = "s"
	NilPrefix       = "n"
	BoolPrefix      = "b"
	JSONPrefix      = "j"
	MapHolderPrefix = "@"
)

var (
	errTypeNotSupport = errors.New("type not support")
	errInvalidData    = errors.New("invalid data")
)

// SerializedJSON type of Serialized json
type SerializedJSON []byte

// Marshal marshal go types to value string
func Marshal(in interface{}) (string, error) {
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
	return "", errTypeNotSupport
}

// MustMarshal marshal go types to value string, panic on error
func MustMarshal(in interface{}) string {
	s, err := Marshal(in)
	if err != nil {
		panic(err)
	}
	return s
}

// Unmarshal unmarshal value string to go types
func Unmarshal(o string) interface{} {
	if len(o) < 1 {
		return errInvalidData
	}
	switch o[0:1] {
	case IntPrefix:
		return rawToInt64(o[1:])
	case StringPrefix:
		return o[1:]
	case NilPrefix:
		return nil
	case BoolPrefix:
		return o[1] == 't'
	case JSONPrefix:
		return SerializedJSON(o[1:])
	case MapHolderPrefix:
		return strings.Split(o, "@")[1:]
	}
	return errInvalidData

}

// MustUnmarshal  unmarshal value string to go types, panic on error
func MustUnmarshal(o string) interface{} {
	rtn := Unmarshal(o)
	if err, ok := rtn.(error); ok {
		panic(err.Error() + ":" + o)
	}
	return rtn
}

func int64ToRaw(i int64) string {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return string(b)
}

func rawToInt64(s string) int64 {
	b := []byte(s)
	return int64(binary.LittleEndian.Uint64(b))
}

func boolToString(i bool) string {
	if i {
		return "t"
	}
	return "f"

}
