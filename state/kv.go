package state

import (
	"github.com/iost-official/prototype/vm"
	"fmt"
	"strconv"
	"github.com/iost-official/prototype/core"
)

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
}

type Key string

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
	t       Type
	val     string
	methods []*vm.Method
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
	return nil
}

func (v *ValueImpl) Decode(bin []byte) error {
	return nil
}

func (v *ValueImpl) Hash() []byte {
	return nil
}

func Merge(a, b Value) (Value, error) {
	if a.Type() != b.Type() {
		return nil, fmt.Errorf("type error")
	}
	return nil, nil
}

func Diff(a, b Value) (Value, error) {
	if a.Type() != b.Type() {
		return nil, fmt.Errorf("type error")
	}
	return nil, nil
}
