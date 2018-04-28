package state

import (
	"fmt"
	"reflect"
	"github.com/iost-official/prototype/common"
	"strconv"
	"encoding/binary"
)

type Key string

type Value interface {
	Type() Type
	String() string

	Encode() []byte
	Decode([]byte) error
	Hash() []byte

	merge(b Value) (Value, error)
	diff(b Value) (Value, error)
}

func Merge(a, b Value) (Value, error) {
	if a == nil || a.Type() == Nil {
		return b, nil
	} else if b == nil || b.Type() == Nil {
		return a, nil
	}
	if a.Type() != b.Type() {
		return nil, fmt.Errorf("type error")
	}
	return a.merge(b)
}

func Diff(a, b Value) (Value, error) {
	if a == nil || a.Type() == Nil {
		return b, nil
	} else if b == nil || b.Type() == Nil {
		return a, nil
	}

	if a.Type() != b.Type() {
		return nil, fmt.Errorf("type error")
	}
	return a.diff(b)
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

var VNil Value = &VInt{
	t: Nil,
	int: 0,
}

type VString struct {
	T Type
	string
}

func MakeVString(s string) VString {
	return VString{
		T:      String,
		string: s,
	}
}
func (v *VString) Type() Type {
	return v.T
}
func (v *VString) String() string {
	return v.string
}
func (v *VString) Encode() []byte {
	raw := ValueRaw{
		t:   uint8(v.T),
		val: []byte(v.string),
	}
	b, err := raw.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return b
}
func (v *VString) Decode(bin []byte) error {
	var raw ValueRaw
	_, err := raw.Unmarshal(bin)
	if err != nil {
		return err
	}
	v.T = Type(raw.t)
	v.string = string(raw.val)
	return nil
}
func (v *VString) Hash() []byte {
	b := v.Encode()
	return common.Sha256(b)
}
func (v *VString) merge(b Value) (Value, error) {
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return nil, fmt.Errorf("type error")
	}
	c := &VString{
		T:      b.Type(),
		string: b.String(),
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
func (v *VString) diff(b Value) (Value, error) {
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return nil, fmt.Errorf("type error")
	}
	c := &VString{
		T:      b.Type(),
		string: b.String(),
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

type VInt struct {
	t Type
	int
}

func MakeVInt(i int) VInt {
	return VInt{Int, i}
}
func (v *VInt) Type() Type {
	return v.t
}
func (v *VInt) String() string {
	return strconv.Itoa(v.int)
}
func (v *VInt) Encode() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(v.int))
	vr := ValueRaw{uint8(v.t), buf}
	ans, err := vr.Marshal(nil)
	if err != nil {
		panic(err)
	}
	return ans
}
func (v *VInt) Decode(bin []byte) error {
	var vr ValueRaw
	_, err := vr.Unmarshal(bin)
	if err != nil {
		return err
	}
	v.int = int(binary.BigEndian.Uint32(bin))
	v.t = Int
	return nil
}
func (v *VInt) Hash() []byte {
	return common.Sha256(v.Encode())
}
func (v *VInt) merge(b Value) (Value, error) {
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return nil, fmt.Errorf("type error")
	}
	vv := reflect.ValueOf(b)
	c := &VInt{
		t:   v.Type(),
		int: vv.Interface().(int),
	}
	return c, nil
}
func (v *VInt) diff(b Value) (Value, error) {
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return nil, fmt.Errorf("type error")
	}
	vv := reflect.ValueOf(b)
	c := &VInt{
		t:   b.Type(),
		int: vv.Interface().(int),
	}
	return c, nil
}

type VBytes struct {
	t   Type
	val []byte
}

func MakeVByte(b []byte) VBytes {
	return VBytes{Bytes, b}
}
func (v *VBytes) Type() Type {
	return v.t
}
func (v *VBytes) String() string {
	return string(v.val)
}
func (v *VBytes) Encode() []byte {
	vr := ValueRaw{uint8(v.t), v.val}
	bin, err := vr.Marshal(nil)
	if err != nil {
		panic(err)
		return nil
	}
	return bin
}
func (v *VBytes) Decode(bin []byte) error {
	var vr ValueRaw
	_, err := vr.Marshal(bin)
	if err != nil {
		return err
	}
	return nil
}
func (v *VBytes) Hash() []byte {
	return common.Sha256(v.Encode())
}
func (v *VBytes) merge(b Value) (Value, error) {
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return nil, fmt.Errorf("type error")
	}
	vv := reflect.ValueOf(b)
	c := &VBytes{
		t:   v.Type(),
		val: vv.Interface().([]byte),
	}
	return c, nil
}
func (v *VBytes) diff(b Value) (Value, error) {
	vv := reflect.ValueOf(b)
	c := &VBytes{
		t:   b.Type(),
		val: vv.Interface().([]byte),
	}
	return c, nil
}
