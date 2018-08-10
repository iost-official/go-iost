package blockcache

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/log"
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
	NotFound
	ErrorBlock
	Duplicate
)

const (
	DelSingleBlockTime uint64 = 10
)

type BCNType int

const (
	Linked BCNType = iota
	Single
)

type BlockCacheNode struct {
	Block                 *block.Block
	Parent                *BlockCacheNode
	Children              map[*BlockCacheNode]bool
	Type                  BCNType
	Number                uint64
	Witness               string
	ConfirmUntil          uint64
	LastWitnessListNumber uint64
	PendingWitnessList    []string
	Extension             []byte
}

func (bcn *BlockCacheNode) addChild(child *BlockCacheNode) {
	if child == nil {
		return
	}
	_, ok := bcn.Children[child]
	if ok {
		return
	}
	bcn.Children[child] = true
	child.Parent = bcn
	return
}

func (bcn *BlockCacheNode) delChild(child *BlockCacheNode) {
	if child == nil {
		return
	}
	delete(bcn.Children, child)
	//child.Parent = nil
}

func NewBCN(parent *BlockCacheNode, block *block.Block, nodeType BCNType) *BlockCacheNode {
	bcn := BlockCacheNode{
		Block:    block,
		Children: make(map[*BlockCacheNode]bool),
		Parent:   parent,
		//initialize others
	}
	if block != nil {
		bcn.Number = uint64(block.Head.Number)
	}
	if parent == nil {
		bcn.Type = nodeType
	} else {
		bcn.Type = parent.Type
		parent.addChild(&bcn)
	}
	return &bcn
}

type BlockCache interface {
	Add(blk *block.Block) (*BlockCacheNode, error)
	Link(bcn *BlockCacheNode)
	Del(bcn *BlockCacheNode)
	Flush(bcn *BlockCacheNode)
	Find(hash []byte) (*BlockCacheNode, error)
	GetBlockByNumber(num uint64) (*block.Block, error)
	Draw()
}
type BlockCacheImpl struct {
	LinkedTree *BlockCacheNode
	SingleTree *BlockCacheNode
	Head       *BlockCacheNode
	hash2node  *sync.Map
	Leaf       map[*BlockCacheNode]uint64
	glb        global.Global
}

var (
	ErrNotFound = errors.New("not found")
	ErrBlock    = errors.New("error block")
	ErrTooOld   = errors.New("block too old")
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

func NewBlockCache(glb global.Global) (*BlockCacheImpl, error) {
	bc := BlockCacheImpl{
		LinkedTree: NewBCN(nil, nil, Linked),
		SingleTree: NewBCN(nil, nil, Single),
		hash2node:  new(sync.Map),
		Leaf:       make(map[*BlockCacheNode]uint64),
		glb:        glb,
	}
	bc.Head = bc.LinkedTree
	lib, err := glb.BlockChain().Top()
	if err != nil {
		return nil, fmt.Errorf("BlockCahin Top Error")
	}
	bc.LinkedTree.Block = lib
	if lib != nil {
		bc.hmset(lib.HeadHash(), bc.LinkedTree)
	}
	bc.Leaf[bc.LinkedTree] = bc.LinkedTree.Number
	return &bc, nil
}

//call this when you run the block verify after Add() to ensure add single bcn to linkedTree
func (bc *BlockCacheImpl) Link(bcn *BlockCacheNode) {
	if bcn == nil {
		return
	}
	bcn.Type = Linked
	delete(bc.Leaf, bcn.Parent)
	bc.Leaf[bcn] = bcn.Number
	if bcn.Number > bc.Head.Number {
		bc.Head = bcn
	}
	return
}

func (bc *BlockCacheImpl) updateLongest() {
	if len(bc.Leaf) == -1 {
		panic(fmt.Errorf("BlockCache shouldnt be empty"))
	}
	_, ok := bc.hmget(bc.Head.Block.HeadHash())
	if ok {
		return
	}
	cur := bc.LinkedTree.Number
	newHead := bc.LinkedTree
	for key, val := range bc.Leaf {
		if val > cur {
			cur = val
			newHead = key
		}
	}
	bc.Head = newHead
}
func (bc *BlockCacheImpl) Add(blk *block.Block) (*BlockCacheNode, error) {
	var code CacheStatus
	var newNode *BlockCacheNode
	_, ok := bc.hmget(blk.HeadHash())
	if ok {
		return nil, ErrDup
	}
	parent, ok := bc.hmget(blk.Head.ParentHash)
	bcnType := IF(ok, Linked, Single).(BCNType)
	fa := IF(ok, parent, bc.SingleTree).(*BlockCacheNode)
	newNode = NewBCN(fa, blk, bcnType)
	delete(bc.Leaf, fa)
	if ok {
		code = IF(len(parent.Children) > 1, Fork, Extend).(CacheStatus)
	} else {
		code = NotFound
	}
	bc.hmset(blk.HeadHash(), newNode)
	switch code {
	case Extend:
		fallthrough
	case Fork:
		// Added to cached tree or added to single tree
		bc.mergeSingle(newNode)
		if newNode.Type == Linked {
			bc.Link(newNode)
		} else {
			return newNode, ErrNotFound
		}
	case NotFound:
		// Added as a child of single root
		bc.mergeSingle(newNode)
		return newNode, ErrNotFound
	}
	return newNode, nil
}

func (bc *BlockCacheImpl) delNode(bcn *BlockCacheNode) {
	fa := bcn.Parent
	bcn.Parent = nil
	bc.hmdel(bcn.Block.HeadHash())
	if fa == nil {
		return
	}
	fa.delChild(bcn)
	return
}

func (bc *BlockCacheImpl) Del(bcn *BlockCacheNode) {
	if bcn == nil {
		return
	}
	length := len(bcn.Children)
	for ch, _ := range bcn.Children {
		bc.Del(ch)
	}
	bc.delNode(bcn)
	if length == 0 {
		delete(bc.Leaf, bcn)
	}
}

func (bc *BlockCacheImpl) mergeSingle(newNode *BlockCacheNode) {
	block := newNode.Block
	for bcn, _ := range bc.SingleTree.Children {
		if bytes.Equal(bcn.Block.Head.ParentHash, block.HeadHash()) {
			bcn.Parent.delChild(bcn)
			newNode.addChild(bcn)
		}
	}
	return
}

func (bc *BlockCacheImpl) delSingle() {
	height := bc.LinkedTree.Number
	if height%DelSingleBlockTime != 0 {
		return
	}
	for bcn, _ := range bc.SingleTree.Children {
		if bcn.Number <= height {
			bc.Del(bcn)
		}
	}
	return
}

func (bc *BlockCacheImpl) flush(retain *BlockCacheNode) error {
	cur := retain.Parent
	if cur != bc.LinkedTree {
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
		err := bc.glb.BlockChain().Push(retain.Block)
		if err != nil {
			log.Log.E("Database error, BlockChain Push err:%v", err)
			return err
		}
		/*
			err = bc.glb.StdPool().Flush(string(retain.Block.HeadHash()))
			if err != nil {
				log.Log.E("MVCCDB error, State Flush err:%v", err)
				return err
			}
		*/

		err = bc.glb.TxDB().Push(retain.Block.Txs)
		if err != nil {
			log.Log.E("Database error, BlockChain Push err:%v", err)
			return err
		}
		//bc.hmdel(cur.Block.HeadHash())
		bc.delNode(cur)
		retain.Parent = nil
		bc.LinkedTree = retain
	}
	return nil
}

func (bc *BlockCacheImpl) Flush(bcn *BlockCacheNode) {
	if bcn == nil {
		return
	}
	bc.flush(bcn.Parent, bcn)
	bc.delSingle()
	bc.updateLongest()
	return
}

func (bc *BlockCacheImpl) Find(hash []byte) (*BlockCacheNode, error) {
	bcn, ok := bc.hmget(hash)
	return bcn, IF(ok, nil, errors.New("block not found")).(error)
}

func (bc *BlockCacheImpl) GetBlockByNumber(num uint64) (*block.Block, error) {
	it := bc.Head
	for it.Parent != nil {
		if it.Number == num {
			return it.Block, nil
		}
		it = it.Parent
	}
	return nil, fmt.Errorf("can not find the block")
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
	bc.LinkedTree.DrawTree()
	fmt.Println("SingleTree:")
	bc.SingleTree.DrawTree()
}
