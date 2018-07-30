package state

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type Key string

type Value interface {
	Type() Type
	EncodeString() string
}

func Merge(a, b Value) Value {

	switch {
	case a == nil:
		panic("Merge from nil, means Get function wrong!")
	case b == nil:
		panic("Merge from nil, means Get function wrong!")
	case b == VNil:
		return a
	case b == VDelete:
		return VNil
	case a.Type() == Map && b.Type() == Map:
		bI := b.(*VMap)
		for k, val := range bI.m {
			a.(*VMap).m[k] = val
		}
		return a
	}

	return b
}

func (k Key) Encode() []byte {
	return []byte(k)
}

type Type int

const (
	Nil Type = iota
	Delete
	Bool
	Int
	Float
	String
	Bytes
	Array // fix length array
	Map
	Stack
	Queue
)

func ParseValue(s string) (Value, error) {

	s1 := string([]rune(s)[1:])

	switch {
	case s == "nil":
		return VNil, nil
	case s == "true":
		return VTrue, nil

	case s == "false ":
		return VFalse, nil
	case strings.HasPrefix(s, "i"):
		i, err := strconv.Atoi(s1)
		if err != nil {
			return nil, err
		}
		return MakeVInt(i), nil
	case strings.HasPrefix(s, "f"):
		f, err := strconv.ParseFloat(s1, 64)
		if err != nil {
			return nil, err
		}
		return MakeVFloat(f), nil

	case strings.HasPrefix(s, "b"):
		b, err := base64.StdEncoding.DecodeString(s1)
		if err != nil {
			return nil, err
		}
		return MakeVByte(b), nil
	case strings.HasPrefix(s, "s"):
		return MakeVString(s[1:]), nil
	case strings.HasPrefix(s, "{"):
		ss := strings.Split(s1, ",")
		if len(ss) <= 0 {
			return MakeVMap(nil), nil
		}
		vmap := MakeVMap(nil)
		for _, kv := range ss {
			if kv == "" {
				continue
			}
			kv1 := strings.Split(kv, ":")
			if len(kv1) != 2 {
				return nil, fmt.Errorf("parsing %v : syntax error", s)
			}
			v, err := ParseValue(kv1[1])
			if err != nil {
				return nil, err
			}
			vmap.Set(Key(kv1[0]), v)
		}
		return vmap, nil
	}
	return nil, fmt.Errorf("parsing %v : syntax error", s)
}

var VNil = &VNilType{}

type VNilType struct{}

func (v *VNilType) Type() Type {
	return Nil
}
func (v *VNilType) EncodeString() string {
	return "nil"
}

var VDelete = &VDeleteType{}

type VDeleteType struct{}

func (v *VDeleteType) Type() Type {
	return Delete
}
func (v *VDeleteType) EncodeString() string {
	return "delete"
}

type VString struct {
	string
}

func MakeVString(s string) *VString {
	return &VString{
		string: s,
	}
}
func (v *VString) Type() Type {
	return String
}
func (v *VString) EncodeString() string {
	return "s" + v.string
}

type VInt struct {
	int
}

func MakeVInt(i int) *VInt {
	return &VInt{
		int: i,
	}
}
func (v *VInt) ToInt() int {
	return v.int
}
func (v *VInt) Type() Type {
	return Int
}
func (v *VInt) EncodeString() string {
	return "i" + strconv.Itoa(v.int)
}

type VBytes struct {
	val []byte
}

func MakeVByte(b []byte) *VBytes {
	return &VBytes{
		val: b,
	}
}

func (v *VBytes) Type() Type {
	return Bytes
}
func (v *VBytes) EncodeString() string {
	return "b" + base64.StdEncoding.EncodeToString(v.val)
}

type VFloat struct {
	float64
}

func MakeVFloat(f float64) *VFloat {
	return &VFloat{
		float64: f,
	}
}

func (v *VFloat) Type() Type {
	return Float
}
func (v *VFloat) EncodeString() string {
	return "f" + strconv.FormatFloat(v.float64, 'e', 15, 64)
}

func (v *VFloat) ToFloat64() float64 {
	return v.float64
}

var VTrue = &VBool{
	val: true,
}

var VFalse = &VBool{
	val: false,
}

type VBool struct {
	val bool
}

func MakeVBool(boo bool) *VBool {
	if boo {
		return VTrue
	} else {
		return VFalse
	}
}

func (v *VBool) Type() Type {
	return Bool
}
func (v *VBool) EncodeString() string {
	if v.val {
		return "true"
	} else {
		return "false"
	}
}

type VMap struct {
	m     map[Key]Value
	mutex sync.RWMutex
}

func MakeVMap(nm map[Key]Value) *VMap {
	if nm == nil {
		nm = make(map[Key]Value)
	}
	return &VMap{
		m: nm,
	}
}

func (v *VMap) Type() Type {
	return Map
}
func (v *VMap) EncodeString() string {
	str := "{"
	for k, val := range v.m {
		str += string(k) + ":" + val.EncodeString() + ","
	}
	return str
}

func (v *VMap) Set(key Key, value Value) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.m[key] = value
}

func (v *VMap) Get(key Key) Value {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	ret, ok := v.m[key]
	if !ok {
		return VNil
	}
	return ret
}

func (v *VMap) Map() map[Key]Value {
	return v.m
}
