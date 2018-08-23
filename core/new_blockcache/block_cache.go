package blockcache

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	blockCachedLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "block_cached_length",
			Help: "Length of cached block chain",
		},
	)
)

func init() {
	prometheus.MustRegister(blockCachedLength)
}

func IF(condition bool, trueRes, falseRes interface{}) interface{} {
	if condition {
		return trueRes
	}
	return falseRes
}

type CacheStatus int

const (
	Extend CacheStatus = iota
	Fork
	ParentNotFound
)

const (
	DelSingleBlockTime int64 = 10
)

type BCNType int

const (
	Linked BCNType = iota
	Single
	Virtual
)

type BlockCacheNode struct {
	Block              *block.Block
	Parent             *BlockCacheNode
	Children           map[*BlockCacheNode]bool
	Type               BCNType
	Number             int64
	Witness            string
	ConfirmUntil       int64
	PendingWitnessList []string
	Extension          []byte
}

func (bcn *BlockCacheNode) addChild(child *BlockCacheNode) {
	if child == nil {
		return
	}
	bcn.Children[child] = true
	child.Parent = bcn
	return
}

func (bcn *BlockCacheNode) delChild(child *BlockCacheNode) {
	delete(bcn.Children, child)
}

func NewBCN(parent *BlockCacheNode, block *block.Block) *BlockCacheNode {
	bcn := BlockCacheNode{
		Block:    block,
		Parent:   parent,
		Children: make(map[*BlockCacheNode]bool),
	}
	if block != nil {
		bcn.Number = block.Head.Number
		bcn.Witness = block.Head.Witness
	}
	if parent != nil {
		bcn.Type = parent.Type
		parent.addChild(&bcn)
	}
	return &bcn
}

type BlockCache interface {
	Add(*block.Block) *BlockCacheNode
	Link(*BlockCacheNode)
	Del(*BlockCacheNode)
	Flush(*BlockCacheNode)
	Find([]byte) (*BlockCacheNode, error)
	GetBlockByNumber(int64) (*block.Block, error)
	GetBlockByHash([]byte) (*block.Block, error)
	LinkedRoot() *BlockCacheNode
	Head() *BlockCacheNode
	Draw()
}

type BlockCacheImpl struct {
	linkedRoot   *BlockCacheNode
	singleRoot   *BlockCacheNode
	head         *BlockCacheNode
	hash2node    *sync.Map
	leaf         map[*BlockCacheNode]int64
	baseVariable global.BaseVariable
}

var (
	ErrNotFound = errors.New("not found")
	ErrDup      = errors.New("block duplicate")
)

func (bc *BlockCacheImpl) hmget(hash []byte) (*BlockCacheNode, bool) {
	rtnI, ok := bc.hash2node.Load(string(hash))
	if !ok {
		return nil, false
	}
	bcn, okn := rtnI.(*BlockCacheNode)
	if !okn {
		bc.hash2node.Delete(string(hash))
		return nil, false
	}
	return bcn, true
}

func (bc *BlockCacheImpl) hmset(hash []byte, bcn *BlockCacheNode) {
	bc.hash2node.Store(string(hash), bcn)
}

func (bc *BlockCacheImpl) hmdel(hash []byte) {
	bc.hash2node.Delete(string(hash))
}

func NewBlockCache(baseVariable global.BaseVariable) (*BlockCacheImpl, error) {
	bc := BlockCacheImpl{
		linkedRoot:   NewBCN(nil, nil),
		singleRoot:   NewBCN(nil, nil),
		hash2node:    new(sync.Map),
		leaf:         make(map[*BlockCacheNode]int64),
		baseVariable: baseVariable,
	}
	bc.linkedRoot.Type = Linked
	bc.singleRoot.Type = Single
	bc.head = bc.linkedRoot
	lib, err := baseVariable.BlockChain().Top()
	if err != nil {
		return nil, fmt.Errorf("BlockCahin Top Error")
	}
	bc.linkedRoot.Block = lib
	if lib != nil {
		bc.hmset(lib.HeadHash(), bc.linkedRoot)
	}
	bc.leaf[bc.linkedRoot] = bc.linkedRoot.Number
	return &bc, nil
}

//call this when you run the block verify after Add() to ensure add single bcn to linkedRoot
func (bc *BlockCacheImpl) Link(bcn *BlockCacheNode) {
	if bcn == nil {
		return
	}
	bcn.Type = Linked
	delete(bc.leaf, bcn.Parent)
	bc.leaf[bcn] = bcn.Number
	if bcn.Number > bc.head.Number {
		bc.head = bcn
	}
	return
}

func (bc *BlockCacheImpl) updateLongest() {
	/*
		think about there are only one witness
		if len(bc.leaf) == -1 {
			panic(fmt.Errorf("BlockCache shouldnt be empty"))
		}
	*/
	_, ok := bc.hmget(bc.head.Block.HeadHash())
	if ok {
		return
	}
	cur := bc.linkedRoot.Number
	for key, val := range bc.leaf {
		if val > cur {
			cur = val
			bc.head = key
		}
	}
}

func (bc *BlockCacheImpl) Add(blk *block.Block) *BlockCacheNode {
	parent, ok := bc.hmget(blk.Head.ParentHash)
	fa := IF(ok, parent, bc.singleRoot).(*BlockCacheNode)
	newNode := NewBCN(fa, blk)
	delete(bc.leaf, fa)
	bc.hmset(blk.HeadHash(), newNode)
	bc.mergeSingle(newNode)
	if newNode.Type == Linked {
		bc.leaf[newNode] = newNode.Number
		if newNode.Number > bc.head.Number {
			bc.head = newNode
		}
	}
	return newNode
}

func (bc *BlockCacheImpl) delNode(bcn *BlockCacheNode) {
	fa := bcn.Parent
	bcn.Parent = nil
	bc.hmdel(bcn.Block.HeadHash())
	if fa == nil {
		return
	}
	fa.delChild(bcn)
}

func (bc *BlockCacheImpl) Del(bcn *BlockCacheNode) {
	if bcn == nil {
		return
	}
	if len(bcn.Children) == 0 {
		delete(bc.leaf, bcn)
	}
	for ch, _ := range bcn.Children {
		bc.Del(ch)
	}
	bc.delNode(bcn)
}

func (bc *BlockCacheImpl) mergeSingle(newNode *BlockCacheNode) {
	for bcn, _ := range bc.singleRoot.Children {
		if bytes.Equal(bcn.Block.Head.ParentHash, newNode.Block.HeadHash()) {
			bc.singleRoot.delChild(bcn)
			newNode.addChild(bcn)
		}
	}
}

func (bc *BlockCacheImpl) delSingle() {
	height := bc.linkedRoot.Number
	if height%DelSingleBlockTime != 0 {
		return
	}
	for bcn, _ := range bc.singleRoot.Children {
		if bcn.Number <= height {
			bc.Del(bcn)
		}
	}
	return
}

func (bc *BlockCacheImpl) flush(retain *BlockCacheNode) error {
	cur := retain.Parent
	if cur != bc.linkedRoot {
		bc.flush(cur)
	}
	for child, _ := range cur.Children {
		if child == retain {
			continue
		}
		bc.Del(child)
	}
	//confirm retain to db
	if retain.Block != nil {
		err := bc.baseVariable.BlockChain().Push(retain.Block)
		if err != nil {
			ilog.Debug("Database error, BlockChain Push err:%v", err)
			return err
		}
		err = bc.baseVariable.StateDB().Flush(string(retain.Block.HeadHash()))
		if err != nil {
			return err
		}

		err = bc.baseVariable.TxDB().Push(retain.Block.Txs)
		if err != nil {
			ilog.Debug("Database error, BlockChain Push err:%v", err)
			return err
		}
		//bc.hmdel(cur.Block.HeadHash())
		bc.delNode(cur)
		retain.Parent = nil
		bc.linkedRoot = retain
	}
	return nil
}

func (bc *BlockCacheImpl) Flush(bcn *BlockCacheNode) {
	bc.flush(bcn)
	bc.delSingle()
	bc.updateLongest()
	return
}

func (bc *BlockCacheImpl) Find(hash []byte) (*BlockCacheNode, error) {
	bcn, ok := bc.hmget(hash)
	if ok {
		return bcn, nil
	} else {
		return nil, errors.New("block not found")
	}
}

func (bc *BlockCacheImpl) GetBlockByNumber(num int64) (*block.Block, error) {
	it := bc.head
	for it.Parent != nil {
		if it.Number == num {
			return it.Block, nil
		}
		it = it.Parent
	}
	return nil, fmt.Errorf("can not find the block")
}

func (bc *BlockCacheImpl) GetBlockByHash(hash []byte) (*block.Block, error) {
	bcn, ok := bc.hmget(hash)
	if !ok {
		return nil, fmt.Errorf("cant find the block")
	}
	return bcn.Block, nil
}

func (bc *BlockCacheImpl) LinkedRoot() *BlockCacheNode {
	return bc.linkedRoot
}

func (bc *BlockCacheImpl) Head() *BlockCacheNode {
	return bc.head
}

//for debug
//draw the blockcache
const PICSIZE int = 100

var pic [PICSIZE][PICSIZE]byte
var picX, picY int

func calcTree(root *BlockCacheNode, x int, y int, isLast bool) int {
	if x > picX {
		picX = x
	}
	if y > picY {
		picY = y
	}
	if y != 0 {
		pic[x][y-1] = '-'
		for i := x; i >= 0; i-- {
			if pic[i][y-2] != ' ' {
				break
			}
			pic[i][y-2] = '|'
		}
	}
	pic[x][y] = 'N'
	var width int = 0
	var f bool = false
	i := 0
	for k, _ := range root.Children {
		if i == len(root.Children)-1 {
			f = true
		}
		width = calcTree(k, x+width, y+2, f)
		i += 1
	}
	if isLast {
		return x + width
	} else {
		return x + width + 2
	}
}

func (bcn *BlockCacheNode) DrawTree() {
	for i := 0; i < PICSIZE; i++ {
		for j := 0; j < PICSIZE; j++ {
			pic[i][j] = ' '
		}
	}
	calcTree(bcn, 0, 0, true)
	for i := 0; i <= picX; i++ {
		for j := 0; j <= picY; j++ {
			fmt.Printf("%c", pic[i][j])
		}
		fmt.Printf("\n")
	}
}

func (bc *BlockCacheImpl) Draw() {
	fmt.Println("\nLinkedTree:")
	bc.linkedRoot.DrawTree()
	fmt.Println("SingleTree:")
	bc.singleRoot.DrawTree()
}
