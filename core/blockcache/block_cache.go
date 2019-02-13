package blockcache

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"os"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/snapshot"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/db/wal"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
	"github.com/xlab/treeprint"
)

var (
	metricsTxTotal = metrics.NewGauge("iost_tx_total", nil)
	metricsDBSize  = metrics.NewGauge("iost_db_size", []string{"Name"})
)

// CacheStatus ...
type CacheStatus int

type conAlgo interface {
	RecoverBlock(blk *block.Block) error
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
	blockCacheWALDir = "./BlockCacheWAL"
)

// BlockCacheNode is the implementation of BlockCacheNode
type BlockCacheNode struct { //nolint:golint
	*block.Block
	rw           sync.RWMutex
	parent       *BlockCacheNode
	Children     map[*BlockCacheNode]bool
	Type         BCNType
	walIndex     uint64
	ValidWitness []string
	WitnessList
	SerialNum int64
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

func encodeUpdateLinkedRootWitness(bc *BlockCacheImpl) (b []byte, err error) {
	uwRaw := &UpdateLinkedRootWitnessRaw{
		BlockHashBytes:    bc.LinkedRoot().HeadHash(),
		LinkedRootWitness: bc.linkedRootWitness,
	}
	b, err = proto.Marshal(uwRaw)
	return
}

func decodeUpdateLinkedRootWitness(b []byte) (blockHeadHash []byte, wt []string, err error) {
	var uwRaw UpdateLinkedRootWitnessRaw
	err = proto.Unmarshal(b, &uwRaw)
	if err != nil {
		return
	}
	blockHeadHash = uwRaw.BlockHashBytes
	wt = uwRaw.LinkedRootWitness
	return
}

func encodeUpdateActive(bcn *BlockCacheNode) (b []byte, err error) {
	// First add block
	uaRaw := &UpdateActiveRaw{
		BlockHashBytes: bcn.HeadHash(),
		WitnessList:    &bcn.WitnessList,
	}
	b, err = proto.Marshal(uaRaw)
	return
}

func decodeUpdateActive(b []byte) (blockHeadHash []byte, wt WitnessList, err error) {
	var uaRaw UpdateActiveRaw
	err = proto.Unmarshal(b, &uaRaw)
	if err != nil {
		return
	}
	blockHeadHash = uaRaw.BlockHashBytes
	wt = *(uaRaw.WitnessList)
	return
}

func encodeBCN(bcn *BlockCacheNode) (b []byte, err error) {
	// First add block
	blockByte, err := bcn.Block.Encode()
	if err != nil {
		return
	}
	bcRaw := &BlockCacheRaw{
		BlockBytes:  blockByte,
		WitnessList: &bcn.WitnessList,
		SerialNum:   bcn.SerialNum,
	}
	b, err = proto.Marshal(bcRaw)
	return
}

func decodeBCN(b []byte) (block block.Block, wt WitnessList, serialNum int64, err error) {
	var bcRaw BlockCacheRaw
	err = proto.Unmarshal(b, &bcRaw)
	if err != nil {
		return
	}
	err = block.Decode(bcRaw.BlockBytes)
	if err != nil {
		return
	}
	wt = *(bcRaw.WitnessList)
	serialNum = bcRaw.SerialNum
	return
}

// NewBCN return a new block cache node instance
func NewBCN(parent *BlockCacheNode, blk *block.Block) *BlockCacheNode {
	bcn := &BlockCacheNode{
		Block:        blk,
		parent:       nil,
		Children:     make(map[*BlockCacheNode]bool),
		ValidWitness: make([]string, 0, 0),
		WitnessList: WitnessList{
			WitnessInfo: make([]string, 0, 0),
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
		parent:       nil,
		Children:     make(map[*BlockCacheNode]bool),
		ValidWitness: make([]string, 0, 0),
		WitnessList: WitnessList{
			WitnessInfo: make([]string, 0, 0),
		},
	}
	if blk != nil {
		bcn.Head.Number = blk.Head.Number - 1
	}
	bcn.setParent(parent)
	bcn.Type = Virtual
	return bcn
}

func (bcn *BlockCacheNode) updateValidWitness() {
	witness := bcn.Head.Witness
	parent := bcn.GetParent()
	for _, w := range parent.ValidWitness {
		bcn.ValidWitness = append(bcn.ValidWitness, w)
		if w == witness {
			witness = ""
			break
		}
	}
	if witness != "" {
		bcn.ValidWitness = append(bcn.ValidWitness, witness)
	}
}

func (bcn *BlockCacheNode) removeValidWitness(root *BlockCacheNode) {
	if bcn != root && bcn.Head.Witness == root.Head.Witness {
		return
	}
	newValidWitness := make([]string, 0, len(bcn.ValidWitness))
	for _, w := range bcn.ValidWitness {
		if w != root.Head.Witness {
			newValidWitness = append(newValidWitness, w)
		}
	}
	bcn.ValidWitness = newValidWitness
	for child := range bcn.Children {
		child.removeValidWitness(root)
	}
}

// BlockCache defines BlockCache's API
type BlockCache interface {
	Add(*block.Block) *BlockCacheNode
	AddGenesis(*block.Block)
	Link(*BlockCacheNode, bool)
	UpdateLib(*BlockCacheNode)
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
	NewWAL(config *common.Config) (err error)
	AddNodeToWAL(bcn *BlockCacheNode)
}

// BlockCacheImpl is the implementation of BlockCache
type BlockCacheImpl struct { //nolint:golint
	linkRW            sync.RWMutex
	linkedRoot        *BlockCacheNode
	singleRoot        *BlockCacheNode
	linkedRootWitness []string
	headRW            sync.RWMutex
	head              *BlockCacheNode
	hash2node         *sync.Map // map[string]*BlockCacheNode
	numberMutex       sync.Mutex
	number2node       *sync.Map // map[int64]*BlockCacheNode
	leaf              map[*BlockCacheNode]int64
	witnessNum        int64
	blockChain        block.Chain
	stateDB           db.MVCCDB
	wal               *wal.WAL
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

func (bc *BlockCacheImpl) nmget(num int64) (*BlockCacheNode, bool) {
	rtnI, ok := bc.number2node.Load(num)
	if !ok {
		return nil, false
	}
	bcn, okn := rtnI.(*BlockCacheNode)
	if !okn {
		bc.number2node.Delete(num)
		return nil, false
	}
	return bcn, true
}

func (bc *BlockCacheImpl) nmset(num int64, bcn *BlockCacheNode) {
	bc.number2node.Store(num, bcn)
}

func (bc *BlockCacheImpl) nmdel(num int64) {
	bc.number2node.Delete(num)
}

// NewBlockCache return a new BlockCache instance
func NewBlockCache(baseVariable global.BaseVariable) (*BlockCacheImpl, error) {
	w, err := wal.Create(baseVariable.Config().DB.LdbPath+blockCacheWALDir, []byte("block_cache_wal"))
	if err != nil {
		return nil, err
	}
	bc := BlockCacheImpl{
		linkedRoot:        NewBCN(nil, nil),
		singleRoot:        NewBCN(nil, nil),
		linkedRootWitness: make([]string, 0),
		hash2node:         new(sync.Map),
		number2node:       new(sync.Map),
		leaf:              make(map[*BlockCacheNode]int64),
		blockChain:        baseVariable.BlockChain(),
		stateDB:           baseVariable.StateDB().Fork(),
		wal:               w,
	}
	bc.linkedRoot.Head.Number = -1

	var lib *block.Block
	if baseVariable.Config().Snapshot.Enable {
		lib, err = snapshot.Load(bc.stateDB)
	} else {
		lib, err = bc.blockChain.Top()
	}

	if err != nil {
		ilog.Errorf("lib not found. err:%v", err)
	}

	ilog.Info("Got LIB: ", lib.Head.Number)
	bc.linkedRoot = NewBCN(nil, lib)
	bc.linkedRoot.Type = Linked
	bc.singleRoot.Type = Virtual
	bc.hmset(bc.linkedRoot.HeadHash(), bc.linkedRoot)
	bc.leaf[bc.linkedRoot] = bc.linkedRoot.Head.Number

	if err := bc.updatePending(bc.linkedRoot); err != nil {
		return nil, err
	}
	bc.LinkedRoot().SetActive(bc.LinkedRoot().Pending()) // For genesis case
	ilog.Info("Witness Block Num:", bc.LinkedRoot().Head.Number)
	for _, v := range bc.linkedRoot.Active() {
		ilog.Info("ActiveWitness:", v)
	}
	for _, v := range bc.linkedRoot.Pending() {
		ilog.Info("PendingWitness:", v)
	}
	bc.head = bc.linkedRoot
	bc.witnessNum = int64(len(bc.LinkedRoot().Pending()))

	return &bc, nil
}

// NewWAL New wal when old one is not recoverable. Move Old File into Corrupted for later analysis.
func (bc *BlockCacheImpl) NewWAL(config *common.Config) (err error) {
	walPath := config.DB.LdbPath + blockCacheWALDir
	corruptWalPath := config.DB.LdbPath + blockCacheWALDir + "Corrupted"
	os.Rename(walPath, corruptWalPath)
	bc.wal, err = wal.Create(walPath, []byte("block_cache_wal"))
	return

}

// Recover recover previews block cache
func (bc *BlockCacheImpl) Recover(p conAlgo) (err error) {
	if bc.wal.HasDecoder() {
		//Get All entries
		_, entries, err := bc.wal.ReadAll()
		if err != nil {
			return err
		}
		ilog.Info("Recover block start")
		for i, entry := range entries {
			if i%2000 == 0 {
				ilog.Infof("Recover block progress:%v/%v", i, len(entries))
			}
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
	case BcMessageType_UpdateActiveType:
		err = bc.applyUpdateActive(bcMessage.Data)
		ilog.Info("Finish ApplySetRoot!")
		if err != nil {
			return
		}
	case BcMessageType_UpdateLinkedRootWitnessType:
		err = bc.applyUpdateLinkedRootWitness(bcMessage.Data)
		if err != nil {
			return
		}
	}
	return
}

func (bc *BlockCacheImpl) applyLink(b []byte, p conAlgo) (err error) {
	block, witnessList, serialNum, err := decodeBCN(b)
	if string(bc.LinkedRoot().HeadHash()) == string(block.HeadHash()) {
		bc.LinkedRoot().WitnessList = witnessList
		bc.LinkedRoot().SerialNum = serialNum
	}
	p.RecoverBlock(&block)
	return err
}

func (bc *BlockCacheImpl) applyUpdateActive(b []byte) (err error) {
	blockHeadHash, witnessList, err := decodeUpdateActive(b)
	if bytes.Equal(blockHeadHash, bc.LinkedRoot().HeadHash()) {
		bc.LinkedRoot().SetActive(witnessList.Active())
		ilog.Infof("Set Root active to :%v", bc.LinkedRoot().Active())
	} else {
		block, ok := bc.hmget(blockHeadHash)
		if ok {
			block.SetActive(witnessList.Active())
		}
	}
	return
}

func (bc *BlockCacheImpl) applyUpdateLinkedRootWitness(b []byte) (err error) {
	blockHeadHash, wl, err := decodeUpdateLinkedRootWitness(b)
	if bytes.Equal(blockHeadHash, bc.LinkedRoot().HeadHash()) {
		bc.linkedRootWitness = wl
		ilog.Infof("Set linkedRootWitness to :%v", bc.linkedRootWitness)
	}
	return
}

// UpdateLib will update last inreversible block
func (bc *BlockCacheImpl) UpdateLib(node *BlockCacheNode) {
	confirmLimit := int(bc.witnessNum*2/3 + 1)

	updateActive := false
	if len(node.ValidWitness) >= confirmLimit {
		bc.updateLib(node, confirmLimit)

		if !common.StringSliceEqual(node.Active(), bc.LinkedRoot().Pending()) {
			updateActive = true
		}
	} else if len(node.ValidWitness)+len(bc.linkedRootWitness) >= confirmLimit &&
		!common.StringSliceEqual(node.Active(), bc.LinkedRoot().Pending()) {
		updateActive = bc.checkUpdateActive(node, confirmLimit)
	}

	if updateActive {
		bc.updateActive(node)
	}
}

func (bc *BlockCacheImpl) updateLib(node *BlockCacheNode, confirmLimit int) {
	if !common.StringSliceEqual(node.Active(), bc.LinkedRoot().Pending()) {
		return
	}
	root := bc.LinkedRoot()
	blockList := make(map[int64]*BlockCacheNode, node.Head.Number-root.Head.Number)
	blockList[node.Head.Number] = node
	loopNode := node.GetParent()
	for loopNode != root {
		blockList[loopNode.Head.Number] = loopNode
		loopNode = loopNode.GetParent()
	}

	for len(node.ValidWitness) >= confirmLimit &&
		common.StringSliceEqual(node.Active(), bc.LinkedRoot().Pending()) &&
		blockList[bc.LinkedRoot().Head.Number+1] != nil {
		// bc.Flush() will change node.ValidWitness, bc.LinkedRoot() and bc.linkedRootWitness
		bc.Flush(blockList[bc.LinkedRoot().Head.Number+1])
	}
}

func (bc *BlockCacheImpl) checkUpdateActive(node *BlockCacheNode, confirmLimit int) bool {
	cnt := len(bc.linkedRootWitness)
	for _, w := range node.ValidWitness {
		inc := true
		for _, w1 := range bc.linkedRootWitness {
			if w == w1 {
				inc = false
				break
			}
		}
		if inc {
			cnt++
		}
		if cnt >= confirmLimit {
			return true
		}
	}
	return false
}

func (bc *BlockCacheImpl) updateActive(node *BlockCacheNode) {
	newValidWitness := make([]string, 0)
	for _, witness := range node.ValidWitness {
		for _, w := range bc.LinkedRoot().Pending() {
			if witness == w {
				newValidWitness = append(newValidWitness, witness)
				break
			}
		}
	}
	node.ValidWitness = newValidWitness
	node.SetActive(bc.LinkedRoot().Pending())
	ilog.Infof("update node:%d activelist to %v", node.Head.Number, node.Active())
	bc.writeUpdateActiveWAL(node)
}

// Link call this when you run the block verify after Add() to ensure add single bcn to linkedRoot
func (bc *BlockCacheImpl) Link(bcn *BlockCacheNode, replay bool) {
	if bcn == nil {
		return
	}
	parent := bcn.GetParent()
	if parent.Type != Linked {
		return
	}
	bcn.Type = Linked
	delete(bc.leaf, parent)
	bc.leaf[bcn] = bcn.Head.Number
	bcn.updateValidWitness()
	bc.updateWitnessList(bcn)
	if !replay {
		bc.AddNodeToWAL(bcn)
	}
	if bcn.Head.Number > bc.Head().Head.Number || (bcn.Head.Number == bc.Head().Head.Number && bcn.Head.Time < bc.Head().Head.Time) {
		bc.SetHead(bcn)
	}
}

// AddNodeToWAL add write node message to WAL
func (bc *BlockCacheImpl) AddNodeToWAL(bcn *BlockCacheNode) {
	index, err := bc.writeAddNodeWAL(bcn)
	if err != nil {
		ilog.Error("Failed to write add node WAL!", err)
	}
	bcn.walIndex = index
}

func (bc *BlockCacheImpl) updateLinkedRootWitness(parent, bcn *BlockCacheNode) {
	if !common.StringSliceEqual(parent.Pending(), bcn.Pending()) {
		bc.linkedRootWitness = make([]string, 0)
	}
	witness := bcn.Head.Witness
	for _, w := range bc.linkedRootWitness {
		if w == witness {
			witness = ""
			break
		}
	}
	if witness != "" {
		bc.linkedRootWitness = append(bc.linkedRootWitness, witness)
	}
}

func (bc *BlockCacheImpl) updateWitnessList(h *BlockCacheNode) error {
	h.CopyWitness(h.GetParent())
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
		return errors.New("failed to checkout state db")
	}
	return nil
}

func (bc *BlockCacheImpl) updateLongest() {
	_, ok := bc.hmget(bc.Head().HeadHash())
	if ok {
		return
	}
	for bcn := range bc.leaf {
		if bcn.Head.Number > bc.Head().Head.Number || (bcn.Head.Number == bc.Head().Head.Number && bcn.Head.Time < bc.Head().Head.Time) {
			bc.SetHead(bcn)
		}
	}
}

// Add is add a block
func (bc *BlockCacheImpl) Add(blk *block.Block) *BlockCacheNode {
	if bc.LinkedRoot().Head.Number >= blk.Head.Number {
		return nil
	}
	newNode, nok := bc.hmget(blk.HeadHash())
	if nok && newNode.Type != Virtual {
		return newNode
	}
	parent, ok := bc.hmget(blk.Head.ParentHash)
	if !ok {
		parent = NewVirtualBCN(bc.singleRoot, blk)
		bc.hmset(blk.Head.ParentHash, parent)
	}
	if nok && newNode.Type == Virtual {
		bc.singleRoot.delChild(newNode)
		newNode.updateVirtualBCN(parent, blk)
	} else {
		newNode = NewBCN(parent, blk)
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

	bc.SetHead(bc.LinkedRoot())
	bc.hmset(bc.LinkedRoot().HeadHash(), bc.LinkedRoot())
	bc.leaf[bc.LinkedRoot()] = bc.LinkedRoot().Head.Number
}

func (bc *BlockCacheImpl) delNode(bcn *BlockCacheNode) {
	parent := bcn.GetParent()
	bcn.SetParent(nil)
	if bcn.Block != nil {
		bc.hmdel(bcn.HeadHash())
	}
	if parent != nil {
		parent.delChild(bcn)
	}
	delete(bc.leaf, bcn)
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
	for ch := range bcn.Children {
		bc.del(ch)
	}
	if bcn.GetParent() != nil && len(bcn.GetParent().Children) == 1 && bcn.GetParent().Type == Linked {
		bc.leaf[bcn.GetParent()] = bcn.GetParent().Head.Number
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

// Flush is save a block
func (bc *BlockCacheImpl) Flush(bcn *BlockCacheNode) {
	parent := bcn.GetParent()
	if parent != bc.LinkedRoot() {
		ilog.Errorf("block isn't blockcache root's child")
	}
	for child := range parent.Children {
		if child == bcn {
			continue
		}
		bc.del(child)
	}
	if bcn.Block == nil {
		ilog.Errorf("When flush, block cache node don't have block: %+v", bcn)
		return
	}

	//confirm bcn to db
	err := bc.blockChain.Push(bcn.Block)
	if err != nil {
		ilog.Errorf("Database error, BlockChain Push err: %v %v", bcn.HeadHash(), err)
	}

	bc.updateLinkedRootWitness(parent, bcn)
	err = bc.writeUpdateLinkedRootWitnessWAL()
	if err != nil {
		ilog.Errorf("write wal error: %v %v", bcn.HeadHash(), err)
	}

	ilog.Debug("confirm: ", bcn.Head.Number)
	err = bc.stateDB.Flush(string(bcn.HeadHash()))

	if err != nil {
		ilog.Errorf("flush mvcc error: %v %v", bcn.HeadHash(), err)
	}

	bcn.removeValidWitness(bcn)
	bc.nmdel(parent.Head.Number)
	bc.delNode(parent)
	bcn.SetParent(nil)
	bc.SetLinkedRoot(bcn)

	metricsTxTotal.Set(float64(bc.blockChain.TxTotal()), nil)

	if blockchainDBSize, err := bc.blockChain.Size(); err != nil {
		ilog.Warnf("Get BlockChainDB size failed: %v", err)
	} else {
		metricsDBSize.Set(
			float64(blockchainDBSize),
			map[string]string{
				"Name": "BlockChainDB",
			},
		)
	}

	if stateDBSize, err := bc.stateDB.Size(); err != nil {
		ilog.Warnf("Get StateDB size failed: %v", err)
	} else {
		metricsDBSize.Set(
			float64(stateDBSize),
			map[string]string{
				"Name": "StateDB",
			},
		)
	}

	bc.delSingle()
	bc.updateLongest()
	bc.cutWALFiles(bcn)
}

func (bc *BlockCacheImpl) writeUpdateLinkedRootWitnessWAL() (err error) {
	hb, err := encodeUpdateLinkedRootWitness(bc)
	if err != nil {
		return err
	}
	bcMessage := &BcMessage{
		Data: hb,
		Type: BcMessageType_UpdateLinkedRootWitnessType,
	}
	data, err := proto.Marshal(bcMessage)
	if err != nil {
		return
	}
	ent := wal.Entry{
		Data: data,
	}
	_, err = bc.wal.SaveSingle(ent)
	return
}

func (bc *BlockCacheImpl) writeUpdateActiveWAL(h *BlockCacheNode) (err error) {
	hb, err := encodeUpdateActive(h)
	if err != nil {
		return err
	}
	bcMessage := &BcMessage{
		Data: hb,
		Type: BcMessageType_UpdateActiveType,
	}
	data, err := proto.Marshal(bcMessage)
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
	bcMessage := &BcMessage{
		Data: hb,
		Type: BcMessageType_LinkType,
	}
	data, err := proto.Marshal(bcMessage)
	if err != nil {
		return 0, err
	}
	ent := wal.Entry{
		Data: data,
	}
	return bc.wal.SaveSingle(ent)
}

func (bc *BlockCacheImpl) cutWALFiles(h *BlockCacheNode) error {
	bc.wal.RemoveFilesBefore(h.walIndex)
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
	bc.numberMutex.Lock()
	for it != nil {
		bcn, ok := bc.nmget(it.Head.Number)
		if ok && bcn == it {
			break
		}
		bc.nmset(it.Head.Number, it)
		it = it.GetParent()
	}
	bc.numberMutex.Unlock()
	bcn, ok := bc.nmget(num)
	if ok {
		return bcn.Block, nil
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
	return linkedTree.String() + "\n" + singleTree.String()
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
