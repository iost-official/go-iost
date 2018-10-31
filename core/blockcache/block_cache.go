package blockcache

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/xlab/treeprint"
)

// CacheStatus ...
type CacheStatus int

const (
	// DelSingleBlockTime ...
	DelSingleBlockTime int64 = 10
)

// BCNType type of BlockCacheNode
type BCNType int

// The types of BlockCacheNode
const (
	Linked BCNType = iota
	Single
	Virtual
)

// BlockCacheNode is the implementation of BlockCacheNode
type BlockCacheNode struct { //nolint:golint
	*block.Block
	WitnessList
	Parent       *BlockCacheNode
	Children     map[*BlockCacheNode]bool
	Type         BCNType
	ConfirmUntil int64
}

func (bcn *BlockCacheNode) addChild(child *BlockCacheNode) {
	if child != nil {
		bcn.Children[child] = true
	}
}

func (bcn *BlockCacheNode) delChild(child *BlockCacheNode) {
	delete(bcn.Children, child)
}

func (bcn *BlockCacheNode) setParent(parent *BlockCacheNode) {
	if parent != nil {
		bcn.Parent = parent
		bcn.Type = Single

		parent.addChild(bcn)
	}
}

func (bcn *BlockCacheNode) updateVirtualBCN(parent *BlockCacheNode, block *block.Block) {
	if bcn.Type == Virtual && parent != nil && block != nil {
		bcn.Block = block
		bcn.setParent(parent)
	}
}

// NewBCN return a new block cache node instance
func NewBCN(parent *BlockCacheNode, blk *block.Block) *BlockCacheNode {
	bcn := &BlockCacheNode{
		Block:    blk,
		Parent:   nil,
		Children: make(map[*BlockCacheNode]bool),
	}
	if blk == nil {
		bcn.Block = &block.Block{
			Head: &block.BlockHead{
				Number: -1,
			},
		}
	}
	bcn.setParent(parent)
	return bcn
}

// NewVirtualBCN return a new virtual block cache node instance
func NewVirtualBCN(parent *BlockCacheNode, blk *block.Block) *BlockCacheNode {
	bcn := &BlockCacheNode{
		Block: &block.Block{
			Head: &block.BlockHead{},
		},
		Parent:   nil,
		Children: make(map[*BlockCacheNode]bool),
	}
	if blk != nil {
		bcn.Head.Number = blk.Head.Number - 1
	}
	bcn.setParent(parent)
	bcn.Type = Virtual
	return bcn
}

// BlockCache defines BlockCache's API
type BlockCache interface {
	Add(*block.Block) *BlockCacheNode
	AddGenesis(*block.Block)
	Link(*BlockCacheNode)
	Del(*BlockCacheNode)
	Flush(*BlockCacheNode)
	Find([]byte) (*BlockCacheNode, error)
	GetBlockByNumber(int64) (*block.Block, error)
	GetBlockByHash([]byte) (*block.Block, error)
	LinkedRoot() *BlockCacheNode
	Head() *BlockCacheNode
	Draw() string
}

// BlockCacheImpl is the implementation of BlockCache
type BlockCacheImpl struct { //nolint:golint
	linkedRoot   *BlockCacheNode
	singleRoot   *BlockCacheNode
	head         *BlockCacheNode
	hash2node    *sync.Map // map[string]*BlockCacheNode
	leaf         map[*BlockCacheNode]int64
	baseVariable global.BaseVariable
	stateDB      db.MVCCDB
}

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

// NewBlockCache return a new BlockCache instance
func NewBlockCache(baseVariable global.BaseVariable) (*BlockCacheImpl, error) {
	bc := BlockCacheImpl{
		linkedRoot:   NewBCN(nil, nil),
		singleRoot:   NewBCN(nil, nil),
		hash2node:    new(sync.Map),
		leaf:         make(map[*BlockCacheNode]int64),
		baseVariable: baseVariable,
		stateDB:      baseVariable.StateDB().Fork(),
	}
	bc.linkedRoot.Head.Number = -1
	lib, err := baseVariable.BlockChain().Top()
	if err == nil {
		bc.linkedRoot = NewBCN(nil, lib)
		bc.linkedRoot.Type = Linked
		bc.singleRoot.Type = Virtual
		bc.hmset(bc.linkedRoot.Block.HeadHash(), bc.linkedRoot)
		bc.leaf[bc.linkedRoot] = bc.linkedRoot.Head.Number

		if err := bc.updatePending(bc.linkedRoot); err != nil {
			return nil, err
		}
		bc.linkedRoot.LibWitnessHandle()
		ilog.Info("Witness Block Num:", bc.LinkedRoot().Head.Number)
		for _, v := range bc.linkedRoot.Active() {
			ilog.Info("ActiveWitness:", v)
		}
		for _, v := range bc.linkedRoot.Pending() {
			ilog.Info("PendingWitness:", v)
		}
	}
	bc.head = bc.linkedRoot
	return &bc, nil
}

// Link call this when you run the block verify after Add() to ensure add single bcn to linkedRoot
func (bc *BlockCacheImpl) Link(bcn *BlockCacheNode) {
	if bcn == nil {
		return
	}
	fa, ok := bc.hmget(bcn.Block.Head.ParentHash)
	if !ok || fa.Type != Linked {
		return
	}
	bcn.Type = Linked
	delete(bc.leaf, bcn.Parent)
	bc.leaf[bcn] = bcn.Head.Number
	bc.setHead(bcn)
	if bcn.Head.Number > bc.head.Head.Number {
		bc.head = bcn
	}
}

func (bc *BlockCacheImpl) setHead(h *BlockCacheNode) error {
	h.CopyWitness(h.Parent)
	if h.Head.Number%common.VoteInterval == 0 {
		if err := bc.updatePending(h); err != nil {
			return err
		}
	}
	return nil
}

func (bc *BlockCacheImpl) updatePending(h *BlockCacheNode) error {

	ok := bc.stateDB.Checkout(string(h.Block.HeadHash()))
	if ok {
		if err := h.UpdatePending(bc.stateDB); err != nil {
			ilog.Error("failed to update pending, err:", err)
			return err
		}
		if err := h.UpdateInfo(bc.stateDB); err != nil {
			ilog.Error("failed to update witness info, err:", err)
			return err
		}
	} else {
		return errors.New("failed to state db")
	}
	return nil
}

func (bc *BlockCacheImpl) updateLongest() {
	_, ok := bc.hmget(bc.head.Block.HeadHash())
	if ok {
		return
	}
	cur := bc.linkedRoot.Head.Number
	for key, val := range bc.leaf {
		if val > cur {
			cur = val
			bc.head = key
		}
	}
}

// Add is add a block
func (bc *BlockCacheImpl) Add(blk *block.Block) *BlockCacheNode {
	newNode, nok := bc.hmget(blk.HeadHash())
	if nok && newNode.Type != Virtual {
		return newNode
	}
	fa, ok := bc.hmget(blk.Head.ParentHash)
	if !ok {
		fa = NewVirtualBCN(bc.singleRoot, blk)
		bc.hmset(blk.Head.ParentHash, fa)
	}
	if nok && newNode.Type == Virtual {
		bc.singleRoot.delChild(newNode)
		newNode.updateVirtualBCN(fa, blk)
	} else {
		newNode = NewBCN(fa, blk)
		bc.hmset(blk.HeadHash(), newNode)
	}
	return newNode
}

// AddGenesis is add genesis block
func (bc *BlockCacheImpl) AddGenesis(blk *block.Block) {
	bc.linkedRoot = NewBCN(nil, blk)
	bc.linkedRoot.Type = Linked

	if err := bc.updatePending(bc.linkedRoot); err == nil {
		bc.linkedRoot.LibWitnessHandle()
	}
	bc.head = bc.linkedRoot
	bc.hmset(bc.linkedRoot.Block.HeadHash(), bc.linkedRoot)
	bc.leaf[bc.linkedRoot] = bc.linkedRoot.Head.Number
}

func (bc *BlockCacheImpl) delNode(bcn *BlockCacheNode) {
	fa := bcn.Parent
	bcn.Parent = nil
	if bcn.Block != nil {
		bc.hmdel(bcn.Block.HeadHash())
	}
	if fa != nil {
		fa.delChild(bcn)
	}
}

// Del is delete a block
func (bc *BlockCacheImpl) Del(bcn *BlockCacheNode) {
	bc.del(bcn)
	bc.updateLongest()
}

func (bc *BlockCacheImpl) del(bcn *BlockCacheNode) {
	if bcn == nil {
		return
	}
	if len(bcn.Children) == 0 {
		delete(bc.leaf, bcn)
	}
	if bcn.Parent != nil && len(bcn.Parent.Children) == 1 && bcn.Parent.Type == Linked {
		bc.leaf[bcn.Parent] = bcn.Parent.Head.Number
	}
	for ch := range bcn.Children {
		bc.del(ch)
	}
	bc.delNode(bcn)
}

func (bc *BlockCacheImpl) delSingle() {
	height := bc.linkedRoot.Head.Number
	if height%DelSingleBlockTime != 0 {
		return
	}
	for bcn := range bc.singleRoot.Children {
		if bcn.Head.Number <= height {
			bc.del(bcn)
		}
	}
}

func (bc *BlockCacheImpl) flush(retain *BlockCacheNode) error {
	cur := retain.Parent
	if cur != bc.linkedRoot {
		bc.flush(cur)
	}
	for child := range cur.Children {
		if child == retain {
			continue
		}
		bc.del(child)
	}
	//confirm retain to db
	if retain.Block != nil {
		err := bc.baseVariable.BlockChain().Push(retain.Block)
		if err != nil {
			ilog.Errorf("Database error, BlockChain Push err:%v", err)
			return err
		}
		ilog.Info("[pob] confirm ", retain.Head.Number)
		err = bc.baseVariable.StateDB().Flush(string(retain.Block.HeadHash()))

		if err != nil {
			ilog.Errorf("flush mvcc error: %v", err)
			return err
		}
		bc.delNode(cur)
		retain.Parent = nil
		retain.LibWitnessHandle()
		bc.linkedRoot = retain
	}
	return nil
}

// Flush is save a block
func (bc *BlockCacheImpl) Flush(bcn *BlockCacheNode) {
	bc.flush(bcn)
	bc.delSingle()
	bc.updateLongest()
}

// Find is find the block
func (bc *BlockCacheImpl) Find(hash []byte) (*BlockCacheNode, error) {
	bcn, ok := bc.hmget(hash)
	if !ok || bcn.Type == Virtual {
		return nil, errors.New("block not found")
	}
	return bcn, nil
}

// GetBlockByNumber get a block by number
func (bc *BlockCacheImpl) GetBlockByNumber(num int64) (*block.Block, error) {
	it := bc.head
	for it != nil {
		if it.Head.Number == num {
			return it.Block, nil
		}
		it = it.Parent
	}
	return nil, fmt.Errorf("block not found")
}

// GetBlockByHash get a block by hash
func (bc *BlockCacheImpl) GetBlockByHash(hash []byte) (*block.Block, error) {
	bcn, err := bc.Find(hash)
	if err != nil {
		return nil, err
	}
	return bcn.Block, nil
}

// LinkedRoot return the root node
func (bc *BlockCacheImpl) LinkedRoot() *BlockCacheNode {
	return bc.linkedRoot
}

// Head return head of block cache
func (bc *BlockCacheImpl) Head() *BlockCacheNode {
	return bc.head
}

// Draw returns the linkedroot's and singleroot's tree graph.
func (bc *BlockCacheImpl) Draw() string {
	linkedTree := treeprint.New()
	bc.linkedRoot.DrawChildren(linkedTree)
	singleTree := treeprint.New()
	bc.singleRoot.DrawChildren(singleTree)
	return linkedTree.String()
}

func (bcn *BlockCacheNode) DrawChildren(root treeprint.Tree) {
	for c := range bcn.Children {
		pattern := strconv.Itoa(int(c.Head.Number))
		if c.Head.Witness != "" {
			pattern += "(" + c.Head.Witness[4:6] + ")"
		}
		root.AddNode(pattern)
		c.DrawChildren(root.FindLastNode())
	}
}
