package blockcache

import (
	"errors"
	"fmt"
	"sync"

	"encoding/json"
	"strconv"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
)

type CacheStatus int

const (
	DelSingleBlockTime int64 = 10
)

type BCNType int

const (
	Linked BCNType = iota
	Single
	Virtual
)

type WitnessList struct {
	activeWitnessList    []string
	pendingWitnessList   []string
	pendingWitnessNumber int64
}

// SetPending set pending witness list
func (wl *WitnessList) SetPending(pl []string) {
	wl.pendingWitnessList = pl
}

// SetPendingNum set block number of pending witness
func (wl *WitnessList) SetPendingNum(n int64) {
	wl.pendingWitnessNumber = n
}

// SetActive set active witness list
func (wl *WitnessList) SetActive(al []string) {
	wl.activeWitnessList = al
}

// Pending get pending witness list
func (wl *WitnessList) Pending() []string {
	return wl.pendingWitnessList
}

// Active get active witness list
func (wl *WitnessList) Active() []string {
	return wl.activeWitnessList
}

// SetPendingNum get block number of pending witness
func (wl *WitnessList) PendingNum() int64 {
	return wl.pendingWitnessNumber
}

// UpdatePending update pending witness list
func (wl *WitnessList) UpdatePending(mv db.MVCCDB) error {

	vi := database.NewVisitor(0, mv)

	spn := database.MustUnmarshal(vi.Get("iost.vote-" + "pendingBlockNumber"))
	// todo delay
	if spn == nil {
		//return errors.New("failed to get pending number")
		return nil
	}
	pn, err := strconv.ParseInt(spn.(string), 10, 64)
	if err != nil {
		return err
	}
	wl.SetPendingNum(pn)

	jwl := database.MustUnmarshal(vi.Get("iost.vote-" + "pendingProducerList"))
	// todo delay
	if jwl == nil {
		//return errors.New("failed to get pending list")
		return nil
	}

	str := make([]string, 0)
	err = json.Unmarshal([]byte(jwl.(string)), &str)
	if err != nil {
		return err
	}
	wl.SetPending(str)

	return nil
}

func (wl *WitnessList) LibWitnessHandle() {
	wl.SetActive(wl.Pending())
}

func (wl *WitnessList) CopyWitness(n *BlockCacheNode) {
	wl.SetActive(n.Active())
	wl.SetPending(n.Pending())
	wl.SetPendingNum(n.PendingNum())
}

type BlockCacheNode struct {
	Block        *block.Block
	Parent       *BlockCacheNode
	Children     map[*BlockCacheNode]bool
	Type         BCNType
	Number       int64
	Witness      string
	ConfirmUntil int64
	WitnessList
	Extension []byte
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
		bcn.Number = block.Head.Number
		bcn.Witness = block.Head.Witness
		bcn.setParent(parent)
	}
}

func NewBCN(parent *BlockCacheNode, block *block.Block) *BlockCacheNode {
	bcn := BlockCacheNode{
		Block:    block,
		Parent:   nil,
		Children: make(map[*BlockCacheNode]bool),
	}
	if block != nil {
		bcn.Number = block.Head.Number
		bcn.Witness = block.Head.Witness
	} else {
		bcn.Number = -1
	}
	bcn.setParent(parent)
	return &bcn
}

func NewVirtualBCN(parent *BlockCacheNode, block *block.Block) *BlockCacheNode {
	bcn := BlockCacheNode{
		Block:    nil,
		Parent:   nil,
		Children: make(map[*BlockCacheNode]bool),
	}
	if block != nil {
		bcn.Number = block.Head.Number - 1
	}
	bcn.setParent(parent)
	bcn.Type = Virtual
	return &bcn
}

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

type BlockCacheImpl struct {
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

func NewBlockCache(baseVariable global.BaseVariable) (*BlockCacheImpl, error) {
	bc := BlockCacheImpl{
		linkedRoot:   NewBCN(nil, nil),
		singleRoot:   NewBCN(nil, nil),
		hash2node:    new(sync.Map),
		leaf:         make(map[*BlockCacheNode]int64),
		baseVariable: baseVariable,
		stateDB:      baseVariable.StateDB().Fork(),
	}
	bc.linkedRoot.Number = -1
	lib, err := baseVariable.BlockChain().Top()
	if err == nil {
		bc.linkedRoot = NewBCN(nil, lib)
		bc.linkedRoot.Type = Linked
		bc.singleRoot.Type = Virtual
		bc.hmset(bc.linkedRoot.Block.HeadHash(), bc.linkedRoot)
		bc.leaf[bc.linkedRoot] = bc.linkedRoot.Number

		if !bc.stateDB.Checkout(string(lib.HeadHash())) {
			return nil, errors.New("failed to state db")
		}
		if err := bc.linkedRoot.UpdatePending(bc.stateDB); err != nil {
			return nil, err
		}

		bc.linkedRoot.LibWitnessHandle()

		ilog.Info("Witness Block Num:", bc.LinkedRoot().Number)
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

//call this when you run the block verify after Add() to ensure add single bcn to linkedRoot
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
	bc.leaf[bcn] = bcn.Number
	if bcn.Number > bc.head.Number {
		bc.setHead(bcn)
	}
}

func (bc *BlockCacheImpl) initHead(h *BlockCacheNode) error {

	bc.head = h

	if err := bc.updatePending(bc.head); err != nil {
		return err
	}

	return nil
}

func (bc *BlockCacheImpl) setHead(h *BlockCacheNode) error {

	h.CopyWitness(bc.head)
	bc.head = h

	if bc.head.Number%common.VoteInterval == 0 {
		if err := bc.updatePending(bc.head); err != nil {
			return err
		}
	}

	return nil
}

func (bc *BlockCacheImpl) updatePending(h *BlockCacheNode) error {

	ok := bc.stateDB.Checkout(string(h.Block.HeadHash()))
	if ok {
		if err := bc.head.UpdatePending(bc.stateDB); err != nil {
			ilog.Error("failed to update pending, err:", err)
			return err
		}
	}

	return nil
}

func (bc *BlockCacheImpl) updateLongest() {
	_, ok := bc.hmget(bc.head.Block.HeadHash())
	if ok {
		return
	}
	cur := bc.linkedRoot.Number
	for key, val := range bc.leaf {
		if val > cur {
			cur = val
			bc.setHead(key)
		}
	}
}

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

func (bc *BlockCacheImpl) AddGenesis(blk *block.Block) {
	bc.linkedRoot = NewBCN(nil, blk)
	bc.linkedRoot.Type = Linked

	if bc.stateDB.Checkout(string(blk.HeadHash())) {
		bc.linkedRoot.UpdatePending(bc.stateDB)
		bc.linkedRoot.LibWitnessHandle()
	}

	bc.head = bc.linkedRoot
	bc.hmset(bc.linkedRoot.Block.HeadHash(), bc.linkedRoot)
	bc.leaf[bc.linkedRoot] = bc.linkedRoot.Number
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
		bc.leaf[bcn.Parent] = bcn.Parent.Number
	}
	for ch, _ := range bcn.Children {
		bc.del(ch)
	}
	bc.delNode(bcn)
}

func (bc *BlockCacheImpl) delSingle() {
	height := bc.linkedRoot.Number
	if height%DelSingleBlockTime != 0 {
		return
	}
	for bcn, _ := range bc.singleRoot.Children {
		if bcn.Number <= height {
			bc.del(bcn)
		}
	}
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
		bc.del(child)
	}
	//confirm retain to db
	if retain.Block != nil {
		err := bc.baseVariable.BlockChain().Push(retain.Block)
		if err != nil {
			ilog.Errorf("Database error, BlockChain Push err:%v", err)
			return err
		}
		ilog.Info("confirm ", retain.Number)
		err = bc.baseVariable.StateDB().Flush(string(retain.Block.HeadHash()))
		if err != nil {
			ilog.Errorf("flush mvcc error: %v", err)
			return err
		}
		err = bc.baseVariable.TxDB().Push(retain.Block.Txs, retain.Block.Receipts)
		if err != nil {
			ilog.Errorf("Database error, Transaction Push err:%v", err)
			return err
		}
		bc.delNode(cur)
		retain.Parent = nil
		retain.LibWitnessHandle()
		bc.linkedRoot = retain
	}
	return nil
}

func (bc *BlockCacheImpl) Flush(bcn *BlockCacheNode) {
	bc.flush(bcn)
	bc.delSingle()
	bc.updateLongest()
}

func (bc *BlockCacheImpl) Find(hash []byte) (*BlockCacheNode, error) {
	bcn, ok := bc.hmget(hash)
	if !ok || bcn.Type == Virtual {
		return nil, errors.New("block not found")
	}
	return bcn, nil
}

func (bc *BlockCacheImpl) GetBlockByNumber(num int64) (*block.Block, error) {
	it := bc.head
	for it != nil {
		if it.Number == num {
			return it.Block, nil
		}
		it = it.Parent
	}
	return nil, fmt.Errorf("block not found")
}

func (bc *BlockCacheImpl) GetBlockByHash(hash []byte) (*block.Block, error) {
	bcn, err := bc.Find(hash)
	if err != nil {
		return nil, err
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

var pic [PICSIZE][PICSIZE]string
var picX, picY int

func calcTree(root *BlockCacheNode, x int, y int, isLast bool) int {
	if x > picX {
		picX = x
	}
	if y > picY {
		picY = y
	}
	if y != 0 {
		pic[x][y-1] = "-"
		for i := x; i >= 0; i-- {
			if pic[i][y-2] != " " {
				break
			}
			pic[i][y-2] = "|"
		}
	}
	pic[x][y] = strconv.FormatInt(root.Number, 10)
	var width int = 0
	var f bool = false
	i := 0
	for k := range root.Children {
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

func (bcn *BlockCacheNode) DrawTree() string {
	var ret string
	for i := 0; i < PICSIZE; i++ {
		for j := 0; j < PICSIZE; j++ {
			pic[i][j] = " "
		}
	}
	calcTree(bcn, 0, 0, true)
	for i := 0; i <= picX; i++ {
		l := ""
		for j := 0; j <= picY; j++ {
			l = l + pic[i][j]
		}
		ret += l
	}
	ilog.Info(ret)
	return ret
}

func (bc *BlockCacheImpl) Draw() string {
	return bc.linkedRoot.DrawTree() + "\n\n" + bc.singleRoot.DrawTree()
}
