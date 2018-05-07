package state

import (
	"encoding/binary"
	"fmt"
	"github.com/iost-official/prototype/common"
	"math"
	"reflect"
	"strconv"
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
	Nil Type = iota
	Bool
	Int
	Float
	String
	Bytes
	Array // wenti
	Map
	MapPatch
	Stack
	StackPatch
	Queue
	QueuePatch
)

var VNil = &VNilType{}

type VNilType struct {}
func (v *VNilType) Type() Type {
	return Nil
}
func (v *VNilType) String() string {
	return ""
}
func (v *VNilType) Encode() []byte {
	return nil
}
func (v *VNilType) Decode([]byte) error {
	return nil
}
func (v *VNilType) Hash() []byte {
	return nil
}
func (v *VNilType) merge(b Value) (Value, error) {
	return nil, nil
}
func (v *VNilType) diff(b Value) (Value, error) {
	return nil, nil
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

func (v *VString) String() string {
	return v.string
}

func (v *VString) Encode() []byte {
	raw := ValueRaw{
		t:   uint8(String),
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
	if raw.t != uint8(String) {
		return fmt.Errorf("type error")
	}
	v.string = string(raw.val)
	return nil
}

func (v *VString) Hash() []byte {
	b := v.Encode()
	return common.Sha256(b)
}

func (v *VString) merge(b Value) (Value, error) {
	// 允许动态类型，下同
	/*
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
	*/

	return b, nil
}
func (v *VString) diff(b Value) (Value, error) {
	/*
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
	*/

	return b, nil
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

func (v *VInt) String() string {
	return strconv.Itoa(v.int)
}

func (v *VInt) Encode() []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(v.int))
	vr := ValueRaw{uint8(Int), buf}
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
	if vr.t != uint8(Int) {
		return fmt.Errorf("type error")
	}
	v.int = int(binary.BigEndian.Uint32(bin))
	return nil
}

func (v *VInt) Hash() []byte {
	return common.Sha256(v.Encode())
}

func (v *VInt) merge(b Value) (Value, error) {
	/*
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return nil, fmt.Errorf("type error")
	}
	vv := reflect.ValueOf(b)
	c := &VInt{
		t:   v.Type(),
		int: vv.Interface().(int),
	}
	return c, nil
	*/

	return b, nil
}

func (v *VInt) diff(b Value) (Value, error) {
	/*
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return nil, fmt.Errorf("type error")
	}
	vv := reflect.ValueOf(b)
	c := &VInt{
		t:   b.Type(),
		int: vv.Interface().(int),
	}
	return c, nil
	*/

	return b, nil
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

func (v *VBytes) String() string {
	return string(v.val)
}

func (v *VBytes) Encode() []byte {
	vr := ValueRaw{uint8(Bytes), v.val}
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
	/*
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return nil, fmt.Errorf("type error")
	}
	vv := reflect.ValueOf(b)
	c := &VBytes{
		t:   v.Type(),
		val: vv.Interface().([]byte),
	}
	return c, nil
	*/

	return b, nil
}

func (v *VBytes) diff(b Value) (Value, error) {
	/*
	vv := reflect.ValueOf(b)
	c := &VBytes{
		t:   b.Type(),
		val: vv.Interface().([]byte),
	}
	return c, nil
	*/

	return b, nil
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

func (v *VFloat) String() string {
	return strconv.FormatFloat(v.float64, 'e', 8, 64)
}

func (v *VFloat) Encode() []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(v.float64))
	return buf[:]
}

func (v *VFloat) Decode(b []byte) error {
	i := binary.BigEndian.Uint64(b)
	v.float64 = math.Float64frombits(i)
	return nil
}

func (v *VFloat) Hash() []byte {
	return common.Sha256(v.Encode())
}

func (v *VFloat) merge(b Value) (Value, error) {
	return b, nil
}

func (v *VFloat) diff(b Value) (Value, error) {
	return b, nil
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

func (v *VBool) String() string {
	if v.val {
		return "true"
	} else {
		return "false"
	}
}

func (v *VBool) Encode() []byte {
	if v.val {
		return []byte{0xff}
	} else {
		return []byte{0}
	}
}

func (v *VBool) Decode(b []byte) error {
	if len(b) < 1 {
		return fmt.Errorf("syntax error")
	}
	if b[0] == 0 {
		v = VFalse
	} else {
		v = VTrue
	}
	return nil
}

func (v *VBool) Hash() []byte {
	return nil
}

func (v *VBool) merge(b Value) (Value, error) {
	return b, nil
}

func (v *VBool) diff(b Value) (Value, error) {
	return b, nil
}

type VMap struct {
	m map[Key]Value
}

type VMapPatch struct {
	deleted_key []Key
	new_key []Key
	new_val []Value
}

func MakeVMap(nm map[Key]Value) *VMap {
	return &VMap{
		m: nm,
	}
}

func MakeVMapPatch(del_k []Key, new_k []Key, new_v []Value) *VMapPatch {
	return &VMapPatch {
		deleted_key: del_k,
		new_key: new_k,
		new_val: new_v,
	}
}

func (v *VMap) Type() Type {
	return Map
}

func (vp *VMapPatch) Type() Type {
	return MapPatch
}

func (v *VMap) String() string {
	str := "{"
	for k, val := range v.m {
		str += string(k) + ":" + val.String() + ";"
	}
	return str + "}"
}

func (vp *VMapPatch) String() string {
	return ""
}

func (v *VMap) Encode() []byte {
	// todo
	return nil
}

func (vp *VMapPatch) Encode() []byte {
	return nil
}

func (v *VMap) Decode([]byte) error {
	// todo
	return nil
}

func (vp *VMapPatch) Decode([]byte) error {
	return nil
}

func (v *VMap) Hash() []byte {
	// todo
	return nil
}

func (vp *VMapPatch) Hash() []byte {
	return nil
}

func (v *VMap) merge(b Value) (Value, error) {
	if reflect.TypeOf(b).Name() != "VMapPatch" {
		return b, nil
	} else {
		b_i := b.(*VMapPatch)
		for _, x := range b_i.deleted_key {
			delete(v.m, x)
		}
		for i, x := range b_i.new_key {
			v.m[x] = b_i.new_val[i]
		}
		return v, nil
	}
}

func (vp *VMapPatch) merge(b Value) (Value, error) {
	return b, nil
}

func (v *VMap) diff(b Value) (Value, error) {
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return b, nil
	}
	b_i := b.(*VMap)
	del_k := []Key{}
	for k, val := range v.m {
		v0, ok := b_i.m[k]
		if !ok || v0 != val {
			del_k = append(del_k, k)
		}
	}
	new_k, new_v := []Key{}, []Value{}
	for k, val := range b_i.m {
		v0, ok := v.m[k]
		if !ok || v0 != val {
			new_k = append(new_k, k)
			new_v = append(new_v, val)
		}
	}
	return MakeVMapPatch(del_k, new_k, new_v), nil
}

func (vp *VMapPatch) diff(b Value) (Value, error) {
	return b, nil
}

func (v *VMap) Set(key Key, value Value) {
	v.m[key] = value
}

func (v *VMap) Get(key Key) (Value, bool) {
	ret, ok := v.m[key]
	return ret, ok
}

const stack_size_limit uint32 = 65536

type VStack struct {
	stk []Value
	top uint32
}

type VStackPatch struct {
	pops uint32
	new_val []Value
}

func MakeVStack(s []Value) *VStack {
	if uint32(len(s)) <= stack_size_limit {
		return &VStack{
			stk: s,
			top: uint32(len(s)),
		}
	} else {
		return &VStack{
			stk: s[:stack_size_limit],
			top: stack_size_limit,
		}
	}
}

func MakeVStackPatch(ps uint32, new_v []Value) *VStackPatch {
	return &VStackPatch {
		pops: ps,
		new_val: new_v,
	}
}

func (v *VStack) Type() Type {
	return Stack
}

func (vp *VStackPatch) Type() Type {
	return StackPatch
}

func (v *VStack) Size() uint32 {
	return v.top
}

func (v *VStack) String() string {
	str := "{[STACK BOTTOM]"
	for i:=uint32(0); i<v.top; i++ {
		str += v.stk[i].String() + ";"
	}
	return str + "[TOP]}"
}

func (vp *VStackPatch) String() string {
	return ""
}

func (v *VStack) Encode() []byte {
	// todo
	return nil
}

func (vp *VStackPatch) Encode() []byte {
	return nil
}

func (v *VStack) Decode([]byte) error {
	// todo
	return nil
}

func (vp *VStackPatch) Decode([]byte) error {
	return nil
}

func (v *VStack) Hash() []byte {
	// todo
	return nil
}

func (vp *VStackPatch) Hash() []byte {
	return nil
}

func (v *VStack) merge(b Value) (Value, error) {
	if reflect.TypeOf(b).Name() != "VStackPatch" {
		return b, nil
	} else {
		b_i := b.(*VStackPatch)
		if v.Size() < b_i.pops {
			return nil, fmt.Errorf("No enough values to pop")
		} else {
			tmp_size := v.Size() - b_i.pops + uint32(len(b_i.new_val))
			if tmp_size <= stack_size_limit {
				v.stk = append(v.stk[:v.Size() - b_i.pops], b_i.new_val...)
				v.top = tmp_size
			} else {
				v.stk = append(v.stk[:v.Size() - b_i.pops], b_i.new_val[:uint32(len(b_i.new_val)) - tmp_size + stack_size_limit]...)
				v.top = stack_size_limit
			}
			return v, nil
		}
	}
}

func (vp *VStackPatch) merge(b Value) (Value, error) {
	return b, nil
}

func (v *VStack) diff(b Value) (Value, error) {
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return b, nil
	}
	b_i, p := b.(*VStack), uint32(0)
	for ; p<v.top && p<b_i.top && v.stk[p]==b_i.stk[p]; p++ {}
	return MakeVStackPatch(v.top - p, b_i.stk[p:]), nil
}

func (vp *VStackPatch) diff(b Value) (Value, error) {
	return b, nil
}

func (v *VStack) Push(val Value) error {
	if v.top == stack_size_limit {
		return fmt.Errorf("Stack size reached limit")
	} else if v.top < uint32(len(v.stk)) {
		v.stk[v.top] = val
	} else {
		v.stk = append(v.stk, val)
	}
	v.top++
	return nil
}

func (v *VStack) Pop() error {
	if v.top > 0 {
		v.top--
		return nil
	} else {
		return fmt.Errorf("Empty stack")
	}
}

const queue_size_limit uint32 = 65536

type VQueue struct {
	q []Value
	front uint32
	rear uint32
}

type VQueuePatch struct {
	outs uint32
	new_val []Value
}

func MakeVQueue(nq []Value) *VQueue {
	if uint32(len(nq)) <= queue_size_limit {
		return &VQueue{
			q: nq,
			front: uint32(0),
			rear: uint32(len(nq)),
		}
	} else {
		return &VQueue{
			q: nq[:queue_size_limit],
			front: uint32(0),
			rear: queue_size_limit,
		}
	}
}

func MakeVQueuePatch(os uint32, new_v []Value) *VQueuePatch {
	return &VQueuePatch {
		outs: os,
		new_val: new_v,
	}
}

func (v *VQueue) Type() Type {
	return Queue
}

func (vp *VQueuePatch) Type() Type {
	return QueuePatch
}

func (v *VQueue) Size() uint32 {
	return v.rear - v.front
}

func (v *VQueue) String() string {
	str := "{[QUEUE FRONT]"
	for i:=v.front; i<v.rear; i++ {
		str += v.q[i].String() + ";"
	}
	return str + "[REAR]}"
}

func (vp *VQueuePatch) String() string {
	return ""
}

func (v *VQueue) Encode() []byte {
	// todo
	return nil
}

func (vp *VQueuePatch) Encode() []byte {
	return nil
}

func (v *VQueue) Decode([]byte) error {
	// todo
	return nil
}

func (vp *VQueuePatch) Decode([]byte) error {
	return nil
}

func (v *VQueue) Hash() []byte {
	// todo
	return nil
}

func (vp *VQueuePatch) Hash() []byte {
	return nil
}

func (v *VQueue) merge(b Value) (Value, error) {
	if reflect.TypeOf(b).Name() != "VQueuePatch" {
		return b, nil
	} else {
		b_i := b.(*VQueuePatch)
		if v.Size() < b_i.outs {
			return nil, fmt.Errorf("No enough values to out")
		} else {
			tmp_size := v.Size() - b_i.outs + uint32(len(b_i.new_val))
			if tmp_size <= queue_size_limit {
				v.front = uint32(0)
				v.q = append(v.q[v.front + b_i.outs:v.rear], b_i.new_val...)
				v.rear = tmp_size
			} else {
				v.front = uint32(0)
				v.q = append(v.q[v.front + b_i.outs:v.rear], b_i.new_val[:uint32(len(b_i.new_val)) - tmp_size + queue_size_limit]...)
				v.rear = queue_size_limit
			}
		}
		return v, nil
	}
}

func (vp *VQueuePatch) merge(b Value) (Value, error) {
	return b, nil
}

func (v *VQueue) diff(b Value) (Value, error) {
	if reflect.TypeOf(b) != reflect.TypeOf(v) {
		return b, nil
	}
	b_i := b.(*VQueue)
	// todo, need KMP algorithm
	return MakeVQueuePatch(v.Size(), b_i.q[b_i.front:b_i.rear]), nil
}

func (vp *VQueuePatch) diff(b Value) (Value, error) {
	return b, nil
}

func (v *VQueue) In(val Value) error {
	if v.Size() == queue_size_limit {
		return fmt.Errorf("Queue size reached limit")
	} else if v.rear < uint32(len(v.q)) {
		v.q[v.rear] = val
	} else {
		v.q = append(v.q, val)
	}
	v.rear++
	if v.rear > queue_size_limit {
		v.q = v.q[v.front:v.rear]
		v.rear -= v.front
		v.front = 0
	}
	return nil
}

func (v *VQueue) Out() error {
	if v.front < v.rear {
		v.front++
		return nil
	} else {
		return fmt.Errorf("Empty queue")
	}
}
