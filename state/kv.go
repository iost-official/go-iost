package state

import (
	"fmt"
	"strconv"
	"github.com/iost-official/prototype/core"
	"reflect"
	"github.com/iost-official/prototype/common"
)

type Key string

type Value interface {
	Type() Type
	// 可以设置为整数，字符串，[]byte
	GetInt() int
	SetInt(v int) error
	GetString() string
	SetString(v string) error
	GetBytes() []byte
	SetBytes(v []byte) error
	core.Serializable

	Merge(b Value) (Value, error)
	Diff(b Value) (Value, error)
}

func Merge(a, b Value) (Value, error) {
	if a.Type() != b.Type() {
		return nil, fmt.Errorf("type error")
	}
	return a.Merge(b)
}

func Diff(a, b Value) (Value, error) {
	if a == nil {
		return b, nil
	} else if b == nil {
		return a, nil
	}

	if a.Type() != b.Type() {
		return nil, fmt.Errorf("type error")
	}
	return a.Diff(b)
}

func (k Key) Encode() []byte {
	return []byte(k)
}

type Type int

const (
	Nil    Type = iota
	Int
	String
	Bytes
	Array
	Map
)

type ValueImpl struct {
	t   Type
	val string
}

func (v *ValueImpl) Type() Type {
	return v.t
}

func (v *ValueImpl) GetInt() int {
	if v.t != Int {
		panic(fmt.Errorf("type error"))
	}
	i, _ := strconv.Atoi(v.val)
	return i
}

func (v *ValueImpl) SetInt(i int) error {
	if v.t == Nil {
		v.t = Int
	} else if v.t != Int {
		return fmt.Errorf("type error")
	}
	v.val = strconv.FormatInt(int64(i), 10)
	return nil
}

func (v *ValueImpl) GetString() string {
	if v.t != Int {
		panic(fmt.Errorf("type error"))
	}
	return v.val
}

func (v *ValueImpl) SetString(s string) error {
	if v.t == Nil {
		v.t = String
	} else if v.t != String {
		return fmt.Errorf("type error")
	}
	v.val = s
	return nil
}

func (v *ValueImpl) GetBytes() []byte {
	if v.t != Bytes {
		panic(fmt.Errorf("type error"))
	}
	return []byte(v.val)
}

func (v *ValueImpl) SetBytes(b []byte) error {
	if v.t == Nil {
		v.t = Bytes
	} else if v.t != Bytes {
		return fmt.Errorf("type error")
	}
	v.val = string(b)
	return nil
}

func (v *ValueImpl) Encode() []byte {
	raw := ValueRaw{
		t:   uint8(v.t),
		val: v.val,
	}
	b, err := raw.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return b
}

func (v *ValueImpl) Decode(bin []byte) error {
	var raw ValueRaw
	_, err := raw.Unmarshal(bin)
	if err != nil {
		return err
	}
	v.t = Type(raw.t)
	v.val = raw.val
	return nil
}

func (v *ValueImpl) Hash() []byte {
	b := v.Encode()
	return common.Sha256(b)
}

func (v *ValueImpl) Merge(b Value) (Value, error) {
	if reflect.TypeOf(b) != reflect.TypeOf(v) ||
		b.Type() != v.Type() {
		return nil, fmt.Errorf("type error")
	}

	c := &ValueImpl{
		t:   b.Type(),
		val: b.GetString(),
	}
	switch v.Type() {
	case Nil:
		return c, nil
	case Int:
		return c, nil
	case String:
		return c, nil
	}

	return c, nil

}

func (v *ValueImpl) Diff(b Value) (Value, error) {
	if reflect.TypeOf(b) != reflect.TypeOf(v) ||
		b.Type() != v.Type() {
		return nil, fmt.Errorf("type error")
	}

	c := &ValueImpl{
		t:   b.Type(),
		val: b.GetString(),
	}

	switch v.Type() {
	case Nil:
		return c, nil
	case Int:
		return c, nil
	case String:
		return c, nil
	}
	return c, nil
}
