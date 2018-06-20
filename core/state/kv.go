package state

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
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
	m map[Key]Value
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
	v.m[key] = value
}

func (v *VMap) Get(key Key) Value {
	ret, ok := v.m[key]
	if !ok {
		return VNil
	}
	return ret
}

func (v *VMap) Map() map[Key]Value {
	return v.m
}

//const stack_size_limit uint32 = 65536
//
//type VStack struct {
//	stk []Value
//	top uint32
//}
//
//type VStackPatch struct {
//	pops    uint32
//	new_val []Value
//}
//
//func MakeVStack(s []Value) *VStack {
//	if uint32(len(s)) <= stack_size_limit {
//		return &VStack{
//			stk: s,
//			top: uint32(len(s)),
//		}
//	} else {
//		return &VStack{
//			stk: s[:stack_size_limit],
//			top: stack_size_limit,
//		}
//	}
//}
//
//func MakeVStackPatch(ps uint32, new_v []Value) *VStackPatch {
//	return &VStackPatch{
//		pops:    ps,
//		new_val: new_v,
//	}
//}
//
//func (v *VStack) Type() Type {
//	return Stack
//}
//
//func (vp *VStackPatch) Type() Type {
//	return StackPatch
//}
//
//func (v *VStack) Size() uint32 {
//	return v.top
//}
//
//func (v *VStack) String() string {
//	str := "{[STACK BOTTOM]"
//	for i := uint32(0); i < v.top; i++ {
//		str += v.stk[i].String() + ";"
//	}
//	return str + "[TOP]}"
//}
//
//func (vp *VStackPatch) String() string {
//	return ""
//}
//
//func (v *VStack) Encode() []byte {
//	return nil
//}
//
//func (vp *VStackPatch) Encode() []byte {
//	return nil
//}
//
//func (v *VStack) Decode([]byte) error {
//	return nil
//}
//
//func (vp *VStackPatch) Decode([]byte) error {
//	return nil
//}
//
//func (v *VStack) Hash() []byte {
//	return nil
//}
//
//func (vp *VStackPatch) Hash() []byte {
//	return nil
//}
//
//func (v *VStack) merge(b Value) (Value, error) {
//	if reflect.TypeOf(b).Name() != "VStackPatch" {
//		return b, nil
//	} else {
//		b_i := b.(*VStackPatch)
//		if v.Size() < b_i.pops {
//			return nil, fmt.Errorf("No enough values to pop")
//		} else {
//			tmp_size := v.Size() - b_i.pops + uint32(len(b_i.new_val))
//			if tmp_size <= stack_size_limit {
//				v.stk = append(v.stk[:v.Size()-b_i.pops], b_i.new_val...)
//				v.top = tmp_size
//			} else {
//				v.stk = append(v.stk[:v.Size()-b_i.pops], b_i.new_val[:uint32(len(b_i.new_val))-tmp_size+stack_size_limit]...)
//				v.top = stack_size_limit
//			}
//			return v, nil
//		}
//	}
//}
//
//func (vp *VStackPatch) merge(b Value) (Value, error) {
//	return b, nil
//}
//
//func (v *VStack) diff(b Value) (Value, error) {
//	if reflect.TypeOf(b) != reflect.TypeOf(v) {
//		return b, nil
//	}
//	b_i, p := b.(*VStack), uint32(0)
//	for ; p < v.top && p < b_i.top && v.stk[p] == b_i.stk[p]; p++ {
//	}
//	return MakeVStackPatch(v.top-p, b_i.stk[p:]), nil
//}
//
//func (vp *VStackPatch) diff(b Value) (Value, error) {
//	return b, nil
//}
//
//func (v *VStack) Push(val Value) error {
//	if v.top == stack_size_limit {
//		return fmt.Errorf("Stack size reached limit")
//	} else if v.top < uint32(len(v.stk)) {
//		v.stk[v.top] = val
//	} else {
//		v.stk = append(v.stk, val)
//	}
//	v.top++
//	return nil
//}
//
//func (v *VStack) Pop() error {
//	if v.top > 0 {
//		v.top--
//		return nil
//	} else {
//		return fmt.Errorf("Empty stack")
//	}
//}
//
//const queue_size_limit uint32 = 65536
//
//type VQueue struct {
//	q     []Value
//	front uint32
//	rear  uint32
//}
//
//type VQueuePatch struct {
//	outs    uint32
//	new_val []Value
//}
//
//func MakeVQueue(nq []Value) *VQueue {
//	if uint32(len(nq)) <= queue_size_limit {
//		return &VQueue{
//			q:     nq,
//			front: uint32(0),
//			rear:  uint32(len(nq)),
//		}
//	} else {
//		return &VQueue{
//			q:     nq[:queue_size_limit],
//			front: uint32(0),
//			rear:  queue_size_limit,
//		}
//	}
//}
//
//func MakeVQueuePatch(os uint32, new_v []Value) *VQueuePatch {
//	return &VQueuePatch{
//		outs:    os,
//		new_val: new_v,
//	}
//}
//
//func (v *VQueue) Type() Type {
//	return Queue
//}
//
//func (vp *VQueuePatch) Type() Type {
//	return QueuePatch
//}
//
//func (v *VQueue) Size() uint32 {
//	return v.rear - v.front
//}
//
//func (v *VQueue) String() string {
//	str := "{[QUEUE FRONT]"
//	for i := v.front; i < v.rear; i++ {
//		str += v.q[i].String() + ";"
//	}
//	return str + "[REAR]}"
//}
//
//func (vp *VQueuePatch) String() string {
//	return ""
//}
//
//func (v *VQueue) Encode() []byte {
//	return nil
//}
//
//func (vp *VQueuePatch) Encode() []byte {
//	return nil
//}
//
//func (v *VQueue) Decode([]byte) error {
//	return nil
//}
//
//func (vp *VQueuePatch) Decode([]byte) error {
//	return nil
//}
//
//func (v *VQueue) Hash() []byte {
//	return nil
//}
//
//func (vp *VQueuePatch) Hash() []byte {
//	return nil
//}
//
//func (v *VQueue) merge(b Value) (Value, error) {
//	if reflect.TypeOf(b).Name() != "VQueuePatch" {
//		return b, nil
//	} else {
//		b_i := b.(*VQueuePatch)
//		if v.Size() < b_i.outs {
//			return nil, fmt.Errorf("No enough values to out")
//		} else {
//			tmp_size := v.Size() - b_i.outs + uint32(len(b_i.new_val))
//			if tmp_size <= queue_size_limit {
//				v.front = uint32(0)
//				v.q = append(v.q[v.front+b_i.outs:v.rear], b_i.new_val...)
//				v.rear = tmp_size
//			} else {
//				v.front = uint32(0)
//				v.q = append(v.q[v.front+b_i.outs:v.rear], b_i.new_val[:uint32(len(b_i.new_val))-tmp_size+queue_size_limit]...)
//				v.rear = queue_size_limit
//			}
//		}
//		return v, nil
//	}
//}
//
//func (vp *VQueuePatch) merge(b Value) (Value, error) {
//	return b, nil
//}
//
//func (v *VQueue) diff(b Value) (Value, error) {
//	if reflect.TypeOf(b) != reflect.TypeOf(v) {
//		return b, nil
//	}
//	b_i := b.(*VQueue)
//	return MakeVQueuePatch(v.Size(), b_i.q[b_i.front:b_i.rear]), nil
//}
//
//func (vp *VQueuePatch) diff(b Value) (Value, error) {
//	return b, nil
//}
//
//func (v *VQueue) In(val Value) error {
//	if v.Size() == queue_size_limit {
//		return fmt.Errorf("Queue size reached limit")
//	} else if v.rear < uint32(len(v.q)) {
//		v.q[v.rear] = val
//	} else {
//		v.q = append(v.q, val)
//	}
//	v.rear++
//	if v.rear > queue_size_limit {
//		v.q = v.q[v.front:v.rear]
//		v.rear -= v.front
//		v.front = 0
//	}
//	return nil
//}
//
//func (v *VQueue) Out() error {
//	if v.front < v.rear {
//		v.front++
//		return nil
//	} else {
//		return fmt.Errorf("Empty queue")
//	}
//}
