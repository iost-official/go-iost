package blockcache

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"os"

	"github.com/gogo/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/db/wal"
	"github.com/iost-official/go-iost/ilog"
	"github.com/xlab/treeprint"
)

// CacheStatus ...
type CacheStatus int

type conAlgo interface {
	RecoverBlock(blk *block.Block, witnessList WitnessList) error
}

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

var (
	blockCacheWALDir = "./block_cache_wal"
)

// BlockCacheNode is the implementation of BlockCacheNode
type BlockCacheNode struct { //nolint:golint
	*block.Block
	rw       sync.RWMutex
	parent   *BlockCacheNode
	Children map[*BlockCacheNode]bool
	Type     BCNType
	walIndex uint64

	ConfirmUntil int64
	WitnessList
}

// GetParent returns the node's parent node.
func (bcn *BlockCacheNode) GetParent() *BlockCacheNode {
	bcn.rw.RLock()
	defer bcn.rw.RUnlock()
	return bcn.parent
}

// SetParent sets the node's parent.
func (bcn *BlockCacheNode) SetParent(p *BlockCacheNode) {
	bcn.rw.Lock()
	bcn.parent = p
	bcn.rw.Unlock()
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
		bcn.SetParent(parent)
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

func encodeBCN(bcn *BlockCacheNode) (b []byte, err error) {
	// First add block
	blockByte, err := bcn.Block.Encode()
	if err != nil {
		return
	}
	bcRaw := BlockCacheRaw{
		BlockBytes:  blockByte,
		WitnessList: &bcn.WitnessList,
	}
	b, err = bcRaw.Marshal()
	return
}
func decodeBCN(b []byte) (block block.Block, witnessList WitnessList, err error) {
	var bcRaw BlockCacheRaw
	err = proto.Unmarshal(b, &bcRaw)
	if err != nil {
		return
	}
	err = block.Decode(bcRaw.BlockBytes)
	if err != nil {
		return
	}
	witnessList = *(bcRaw.WitnessList)
	return
}

// NewBCN return a new block cache node instance
func NewBCN(parent *BlockCacheNode, blk *block.Block) *BlockCacheNode {
	bcn := &BlockCacheNode{
		Block:    blk,
		parent:   nil,
		Children: make(map[*BlockCacheNode]bool),
		WitnessList: WitnessList{
			WitnessInfo: make(map[string]*WitnessInfo),
		},
	}
	if blk == nil {
		bcn.Block = &block.Block{
			Head: &block.BlockHead{
				Number: -1,
			}}
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
		parent:   nil,
		Children: make(map[*BlockCacheNode]bool),
		WitnessList: WitnessList{
			WitnessInfo: make(map[string]*WitnessInfo),
		},
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
	AddWithWit(blk *block.Block, witnessList WitnessList) (bcn *BlockCacheNode)
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
	CleanDir() error
	Recover(p conAlgo) (err error)
}

// BlockCacheImpl is the implementation of BlockCache
type BlockCacheImpl struct { //nolint:golint
	linkRW       sync.RWMutex
	linkedRoot   *BlockCacheNode
	singleRoot   *BlockCacheNode
	headRW       sync.RWMutex
	head         *BlockCacheNode
	hash2node    *sync.Map // map[string]*BlockCacheNode
	leaf         map[*BlockCacheNode]int64
	baseVariable global.BaseVariable
	stateDB      db.MVCCDB
	wal          *wal.WAL
}

// CleanDir used in test to clean dir
func (bc *BlockCacheImpl) CleanDir() error {
	if bc.wal != nil {
		return bc.wal.CleanDir()
	}
	return nil
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
	w, err := wal.Create(blockCacheWALDir, []byte("block_cache_wal"))
	if err != nil {
		return nil, err
	}
	bc := BlockCacheImpl{
		linkedRoot:   NewBCN(nil, nil),
		singleRoot:   NewBCN(nil, nil),
		hash2node:    new(sync.Map),
		leaf:         make(map[*BlockCacheNode]int64),
		baseVariable: baseVariable,
		stateDB:      baseVariable.StateDB().Fork(),
		wal:          w,
	}
	bc.linkedRoot.Head.Number = -1
	lib, err := baseVariable.BlockChain().Top()
	if err == nil {
		bc.linkedRoot = NewBCN(nil, lib)
		bc.linkedRoot.Type = Linked
		bc.singleRoot.Type = Virtual
		bc.hmset(bc.linkedRoot.HeadHash(), bc.linkedRoot)
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

// Recover recover previews block cache
func (bc *BlockCacheImpl) Recover(p conAlgo) (err error) {
	if bc.wal.HasDecoder() {
		//Get All entries
		_, entries, err := bc.wal.ReadAll()
		if err != nil {
			return err
		}
		for _, entry := range entries {
			err := bc.apply(entry, p)
			if err != nil {
				return err
			}
		}
	}
	return
}

func (bc *BlockCacheImpl) apply(entry wal.Entry, p conAlgo) (err error) {
	var bcMessage BcMessage
	proto.Unmarshal(entry.Data, &bcMessage)
	switch bcMessage.Type {
	case BcMessageType_LinkType:
		err = bc.applyLink(bcMessage.Data, p)
		if err != nil {
			return
		}
	case BcMessageType_SetRootType:
		err = bc.applySetRoot(bcMessage.Data)
		if err != nil {
			return
		}
	}
	return
}

func (bc *BlockCacheImpl) applyLink(b []byte, p conAlgo) (err error) {
	block, witnessList, err := decodeBCN(b)
	//bc.Add(&block)
	p.RecoverBlock(&block, witnessList)

	return
}

func (bc *BlockCacheImpl) applySetRoot(b []byte) (err error) {

	return
}

// Link call this when you run the block verify after Add() to ensure add single bcn to linkedRoot
func (bc *BlockCacheImpl) Link(bcn *BlockCacheNode) {
	if bcn == nil {
		return
	}
	fa, ok := bc.hmget(bcn.Head.ParentHash)
	if !ok || fa.Type != Linked {
		return
	}
	index, err := bc.writeAddNodeWAL(bcn)
	if err != nil {
		ilog.Error("Failed to write add node WAL!")
	}
	bcn.walIndex = index
	bcn.Type = Linked
	delete(bc.leaf, bcn.GetParent())
	bc.leaf[bcn] = bcn.Head.Number
	bc.setHead(bcn)
	if bcn.Head.Number > bc.Head().Head.Number {
		bc.SetHead(bcn)
	}
}

func (bc *BlockCacheImpl) setHead(h *BlockCacheNode) error {
	if h.PendingWitnessNumber == 0 && h.Active() == nil && h.Pending() == nil {
		h.CopyWitness(h.GetParent())
	}
	if h.Head.Number%common.VoteInterval == 0 {
		if err := bc.updatePending(h); err != nil {
			return err
		}
	}
	return nil
}

func (bc *BlockCacheImpl) updatePending(h *BlockCacheNode) error {

	ok := bc.stateDB.Checkout(string(h.HeadHash()))
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
	_, ok := bc.hmget(bc.Head().HeadHash())
	if ok {
		return
	}
	cur := bc.LinkedRoot().Head.Number
	for key, val := range bc.leaf {
		if val > cur {
			cur = val
			bc.SetHead(key)
		}
	}
}

// AddWithWit add block with witnessList
func (bc *BlockCacheImpl) AddWithWit(blk *block.Block, witnessList WitnessList) (bcn *BlockCacheNode) {
	bcn = bc.Add(blk)
	bcn.WitnessList = witnessList
	return
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
	//newNode.WitnessInfo = wi
	return newNode
}

// AddGenesis is add genesis block
func (bc *BlockCacheImpl) AddGenesis(blk *block.Block) {
	l := NewBCN(nil, blk)
	l.Type = Linked
	bc.SetLinkedRoot(l)

	if err := bc.updatePending(bc.LinkedRoot()); err == nil {
		bc.LinkedRoot().LibWitnessHandle()
	}
	bc.SetHead(bc.LinkedRoot())
	bc.hmset(bc.LinkedRoot().HeadHash(), bc.LinkedRoot())
	bc.leaf[bc.LinkedRoot()] = bc.LinkedRoot().Head.Number
}

func (bc *BlockCacheImpl) delNode(bcn *BlockCacheNode) {
	fa := bcn.GetParent()
	bcn.SetParent(nil)
	if bcn.Block != nil {
		bc.hmdel(bcn.HeadHash())
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
	if bcn.GetParent() != nil && len(bcn.GetParent().Children) == 1 && bcn.GetParent().Type == Linked {
		bc.leaf[bcn.GetParent()] = bcn.GetParent().Head.Number
	}
	for ch := range bcn.Children {
		bc.del(ch)
	}
	bc.delNode(bcn)
}

func (bc *BlockCacheImpl) delSingle() {
	height := bc.LinkedRoot().Head.Number
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
	cur := retain.GetParent()
	if cur != bc.LinkedRoot() {
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
		ilog.Debug("confirm: ", retain.Head.Number)
		err = bc.baseVariable.StateDB().Flush(string(retain.HeadHash()))

		if err != nil {
			ilog.Errorf("flush mvcc error: %v", err)
			return err
		}
		bc.delNode(cur)
		retain.SetParent(nil)
		retain.LibWitnessHandle()
		bc.SetLinkedRoot(retain)
	}
	return nil
}

// Flush is save a block
func (bc *BlockCacheImpl) Flush(bcn *BlockCacheNode) {
	bc.flush(bcn)
	bc.delSingle()
	bc.updateLongest()
	bc.flushWAL(bcn)
}

func (bc *BlockCacheImpl) flushWAL(h *BlockCacheNode) error {
	err := bc.writeSetHeadWAL(h)
	if err != nil {
		return err
	}
	err = bc.cutWALFiles(h)
	if err != nil {
		return err
	}
	return nil
}

func (bc *BlockCacheImpl) writeSetHeadWAL(h *BlockCacheNode) (err error) {
	bcMessage := BcMessage{
		Data: h.Block.HeadHash(),
		Type: BcMessageType_SetRootType,
	}
	data, err := bcMessage.Marshal()
	if err != nil {
		return
	}
	ent := wal.Entry{
		Data: data,
	}
	_, err = bc.wal.SaveSingle(ent)
	return
}

func (bc *BlockCacheImpl) writeAddNodeWAL(h *BlockCacheNode) (uint64, error) {
	hb, err := encodeBCN(h)
	if err != nil {
		return 0, err
	}
	bcMessage := BcMessage{
		Data: hb,
		Type: BcMessageType_LinkType,
	}
	data, err := bcMessage.Marshal()
	if err != nil {
		return 0, err
	}
	ent := wal.Entry{
		Data: data,
	}
	return bc.wal.SaveSingle(ent)
}
func (bc *BlockCacheImpl) cutWALFiles(h *BlockCacheNode) error {
	bc.wal.RemoveFiles(h.walIndex)
	return nil
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
	it := bc.Head()
	if num < bc.LinkedRoot().Head.Number || num > it.Head.Number {
		return nil, fmt.Errorf("block not found")
	}
	for it != nil {
		if it.Head.Number == num {
			return it.Block, nil
		}
		it = it.GetParent()
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
	bc.linkRW.RLock()
	defer bc.linkRW.RUnlock()
	return bc.linkedRoot
}

// SetLinkedRoot sets linked blockcache node.
func (bc *BlockCacheImpl) SetLinkedRoot(n *BlockCacheNode) {
	bc.linkRW.Lock()
	bc.linkedRoot = n
	bc.linkRW.Unlock()
}

// Head return head of block cache
func (bc *BlockCacheImpl) Head() *BlockCacheNode {
	bc.headRW.RLock()
	defer bc.headRW.RUnlock()
	return bc.head
}

// SetHead sets head blockcache node.
func (bc *BlockCacheImpl) SetHead(n *BlockCacheNode) {
	bc.headRW.Lock()
	bc.head = n
	bc.headRW.Unlock()
}

// Draw returns the linkedroot's and singleroot's tree graph.
func (bc *BlockCacheImpl) Draw() string {
	linkedTree := treeprint.New()
	bc.LinkedRoot().drawChildren(linkedTree)
	singleTree := treeprint.New()
	bc.singleRoot.drawChildren(singleTree)
	return linkedTree.String()
}

func (bcn *BlockCacheNode) drawChildren(root treeprint.Tree) {
	for c := range bcn.Children {
		pattern := strconv.Itoa(int(c.Head.Number))
		if c.Head.Witness != "" {
			pattern += "(" + c.Head.Witness[4:6] + ")"
		}
		root.AddNode(pattern)
		c.drawChildren(root.FindLastNode())
	}
}

// CleanBlockCacheWAL used in test to clean dir
func CleanBlockCacheWAL() error {
	return os.RemoveAll(blockCacheWALDir)
}
