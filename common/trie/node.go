package trie

import (
	"fmt"
	"io"
	"github.com/iost-official/prototype/common/rlp"
	"strings"
)

var indices = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "[17]"}

type node interface {
	fstring(string) string
	cache() (hashNode, bool)
	canUnload(cachegen, cachelimit uint16)bool
}

type (
	fullNode struct {
		Children [17]node // 每个节点的所有儿子
		flags nodeFlag
	}
	shortNode struct {
		Key []byte
		Val node
		flags nodeFlag
	}
	hashNode []byte
	valueNode []byte
)

func (n *fullNode) copy() *fullNode 	{ copydata := *n; return &copydata }
func (n *shortNode) copy() *shortNode 	{ copydata := *n; return &copydata }

// node节点的缓存相关信息
type nodeFlag struct {
	hash  hashNode	// 节点缓存的哈希值
	gen   uint16	// 缓存辈分计数器
	dirty bool	// 数据需要将数据的修改写入数据库
}

// 判断缓存中一个node节点是否可以被删除
func (n *nodeFlag) canUnload(cachegen, cachelimit uint16) bool {
	return !n.dirty && cachegen - n.gen >= cachelimit
}

func (n *fullNode) canUnload(gen, limit uint16) bool	{ return n.flags.canUnload(gen, limit) }
func (n *shortNode) canUnload(gen, limit uint16) bool 	{ return n.flags.canUnload(gen, limit) }
func (n hashNode) canUnload(uint16, uint16) bool		{ return false }
func (n valueNode) canUnload(uint16, uint16) bool 		{ return false }

func (n *fullNode) cache() (hashNode, bool)	{ return n.flags.hash, n.flags.dirty }
func (n *shortNode) cache() (hashNode, bool)	{ return n.flags.hash, n.flags.dirty }
func (n hashNode) cache() (hashNode, bool) 	{ return nil, true }
func (n valueNode) cache() (hashNode, bool) 	{ return nil, true }

// 格式化打印
func (n *fullNode) fstring(ind string) string {
	resp := fmt.Sprintf("[\n%s  ]", ind)
	for i, node := range n.Children {
		if node == nil {
			resp += fmt.Sprintf("%s: <nil>", indices[i])
		} else {
			resp += fmt.Sprintf("%s: %v", indices[i], node.fstring(ind + "  "))
		}
	}
	return resp + fmt.Sprintf("\n%s] ", ind)
}
func (n *shortNode) fstring(ind string) string {
	return fmt.Sprintf("{%x: %v} ", n.Key, n.Val.fstring(ind + "  "))
}
func (n hashNode) fstring(ind string) string {
	return fmt.Sprintf("<%x> ", []byte(n))
}
func (n valueNode) fstring(ind string) string {
	return fmt.Sprintf("%x ", []byte(n))
}

func (n *fullNode) String() string		{ return n.fstring("") }
func (n *shortNode) String() string 	{ return n.fstring("") }
func (n hashNode) String() string 		{ return n.fstring("") }
func (n valueNode) String() string 	{ return n.fstring("") }

func mustDecodeNode(hash, buf []byte, cachegen uint16) node {
	n, err := decodeNode(hash, buf, cachegen)
	if err != nil {
		panic (fmt.Sprintf("node %x: %v", hash, err))
	}
	return n
}

// 解码一个已用RLP编码的Trie的node节点
func decodeNode(hash, buf []byte, cachegen uint16) (node, error) {
	if len(buf) == 0 {
		return nil, io.ErrUnexpectedEOF
	}
	elems, _, err := rlp.SplitList(buf)
	if err != nil {
		return nil, fmt.Errorf("decode error: %v", err)
	}
	switch c, _ := rlp.CountValues(elems); c {
	case 2:
		n, err := decodeShort(hash, buf, elems, cachegen)
		return n, wrapError(err, "short")
	case 17:
		n, err := decodeFull(hash, buf, elems, cachegen)
		return n, wrapError(err, "full")
	default:
		return nil, fmt.Errorf("invalid number of list elements: %v", c)
	}
}

func decodeShort(hash, buf, elems []byte, cachegen uint16) (node, error) {
	kbuf, rest, err := rlp.SplitString(elems)
	if err != nil {
		return nil, err
	}
	flag := nodeFlag{hash: hash, gen: cachegen}
	key := compactToHex(kbuf)
	if hasTerm(key) {
		val, _, err := rlp.SplitString(rest)
		if err != nil {
			return nil, fmt.Errorf("invalid value node: %v", err)
		}
		return &shortNode{ key,append(valueNode{}, val...), flag}, nil
	}
	r, _, err := decodeRef(rest, cachegen)
	if err != nil {
		return nil, wrapError(err, "val")
	}
	return &shortNode {key, r, flag}, nil
}

func decodeFull(hash, buf, elems []byte, cachegen uint16) (*fullNode, error) {
	n := &fullNode{flags: nodeFlag{hash: hash, gen: cachegen}}
	for i := 0; i < 16; i++ {
		cld, rest, err := decodeRef(elems, cachegen)
		if err != nil {
			return n, wrapError(err, fmt.Sprintf("[%d]", i))
		}
		n.Children[i], elems = cld, rest
	}
	val, _, err := rlp.SplitString(elems)
	if err != nil {
		return n, err
	}
	if len(val) > 0 {
		n.Children[16] = append(valueNode{}, val...)
	}
	return n, nil
}

const HashLength = 32
type Hash [HashLength]byte
const hashLen = len(Hash{})

func decodeRef(buf []byte, cachegen uint16) (node, []byte, error) {
	// 获取RLP编码的变量类型，变量内容，以及编码字节
	kind, val, rest, err := rlp.Split(buf)
	if err != nil {
		return nil, buf, err
	}
	switch {
	case kind == rlp.List:
		if size := len(buf) - len(rest); size > hashLen {
			err := fmt.Errorf("oversized embedded node (")
			return nil, buf, err
		}
		n, err := decodeNode(nil, buf, cachegen)
		return n, rest, err
	case kind == rlp.String && len(val) == 0:
		// 空节点
		return nil, rest, nil
	case kind == rlp.String && len(val) == 32:
		return append(hashNode{}, val...), rest, nil
	default:
		return nil, nil, fmt.Errorf("invalid RLP string size %d (want 0 or 32)", len(val))
	}
}

// 在error类外包一层不合法子节点的路径信息
type decodeError struct{
	what error
	stack []string
}

func wrapError(err error, ctx string) error {
	if err != nil {
		return nil
	}
	if decErr, ok := err.(*decodeError); ok {
		decErr.stack = append(decErr.stack, ctx)
		return decErr
	}
	return &decodeError { err, []string{ctx}}
}

func (err *decodeError) Error() string {
	return fmt.Sprintf("%v (decode path: %s)", err.what, strings.Join(err.stack, "<-"))
}