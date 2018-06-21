package blockcache

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/log"
	//"github.com/iost-official/prototype/log"
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

//const (
//	MaxCacheDepth = 6
//)

// CacheStatus 代表缓存块的状态
type CacheStatus int

const (
	Extend     CacheStatus = iota // 链增长
	Fork                          // 分叉
	NotFound                      // 无法上链，成为孤块
	ErrorBlock                    // 块有错误
	Duplicate                     // 重复块
)

type BCTType int

const (
	OnCache BCTType = iota
	Singles
)

// BlockCacheTree 缓存链分叉的树结构
type BlockCacheTree struct {
	bc       CachedBlockChain
	children []*BlockCacheTree
	super    *BlockCacheTree
	pool     state.Pool
	bctType  BCTType
}

func newBct(block *block.Block, tree *BlockCacheTree) *BlockCacheTree {
	bct := BlockCacheTree{
		bc:       tree.bc.Copy(),
		children: make([]*BlockCacheTree, 0),
		super:    tree,
		bctType:  Singles,
	}
	bct.bc.Push(block)
	return &bct
}

/*
func (b *BlockCacheTree) add(block *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) (CacheStatus, *BlockCacheTree) {
	if bytes.Equal(b.bc.Top().HeadHash(), block.Head.ParentHash) {
		bct := newBct(block, b)
		if b.bctType == OnCache {
			newPool, err := verifier(block, b.bc.Top(), b.pool)
			if err != nil {
				log.Log.I("ErrorBlock: %v\n", err)
				return ErrorBlock, nil
			}
			bct.pool = newPool
		}
		b.children = append(b.children, bct)
		if len(b.children) == 1 {
			return Extend, bct
		} else {
			return Fork, bct
		}
	}
	return NotFound, nil
}
*/

func (b *BlockCacheTree) popLongest() *BlockCacheTree {
	for _, bct := range b.children {
		if bct.bc.depth == b.bc.depth-1 {
			return bct
		}
	}
	return nil
}

func (b *BlockCacheTree) updateLength() {
	for _, bct := range b.children {
		if bct.bc.parent == &b.bc {
			bct.bc.cachedLength = b.bc.cachedLength + 1
		}
		bct.updateLength()
	}
}

func (b *BlockCacheTree) iterate(fun func(bct *BlockCacheTree) bool) bool {
	if fun(b) {
		return true
	}
	for _, bct := range b.children {
		f := bct.iterate(fun)
		if f {
			return true
		}
	}
	return false
}

var (
	ErrNotFound = errors.New("not found")       // 没有上链，成为孤块
	ErrBlock    = errors.New("error block")     // 块有错误
	ErrTooOld   = errors.New("block too old")   // 块太老
	ErrDup      = errors.New("block duplicate") // 重复块
)

// blockCache 操作块缓存的接口
type BlockCache interface {
	AddGenesis(block *block.Block) error
	Add(block *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) error

	FindBlockInCache(hash []byte) (*block.Block, error)
	CheckBlock(hash []byte) bool
	LongestChain() block.Chain
	LongestPool() state.Pool
	BlockChain() block.Chain
	BasePool() state.Pool
	SetBasePool(statePool state.Pool) error
	ConfirmedLength() uint64
	BlockConfirmChan() chan uint64
	OnBlockChan() chan *block.Block
	SendOnBlock(blk *block.Block)
	Draw()
}

// BlockCacheImpl 块缓存实现
type BlockCacheImpl struct {
	bc                 block.Chain
	cachedRoot         *BlockCacheTree
	singleBlockRoot    *BlockCacheTree
	hashMap            map[string]*BlockCacheTree
	hmlock             sync.RWMutex
	maxDepth           int
	blkConfirmChan     chan uint64
	chConfirmBlockData chan *block.Block
}

// NewBlockCache 新建块缓存
// chain 已确认链部分, pool 已确认状态池, maxDepth 和共识相关的确认块参数
func NewBlockCache(chain block.Chain, pool state.Pool, maxDepth int) *BlockCacheImpl {
	h := BlockCacheImpl{
		bc: chain,
		cachedRoot: &BlockCacheTree{
			bc:       NewCBC(chain),
			children: make([]*BlockCacheTree, 0),
			super:    nil,
			pool:     pool,
			bctType:  OnCache,
		},
		singleBlockRoot: &BlockCacheTree{
			bc:       NewCBC(chain),
			children: make([]*BlockCacheTree, 0),
			super:    nil,
			bctType:  Singles,
		},
		hashMap:            make(map[string]*BlockCacheTree),
		maxDepth:           maxDepth,
		blkConfirmChan:     make(chan uint64, 10),
		chConfirmBlockData: make(chan *block.Block, 100),
	}
	if h.cachedRoot.bc.Top() != nil {
		h.hashMap[string(h.cachedRoot.bc.Top().HeadHash())] = h.cachedRoot
	}
	return &h
}

// ConfirmedLength 返回确认链长度
func (h *BlockCacheImpl) ConfirmedLength() uint64 {
	return h.bc.Length()
}

// BlockChain 返回确认
func (h *BlockCacheImpl) BlockChain() block.Chain {
	return h.bc
}

// AddGenesis 加入创世块
func (h *BlockCacheImpl) AddGenesis(block *block.Block) error {

	err := h.bc.Push(block)
	if err != nil {
		return err
	}

	err = h.cachedRoot.pool.Flush()
	if err != nil {
		return err
	}
	h.hashMap[string(h.cachedRoot.bc.Top().HeadHash())] = h.cachedRoot
	return nil
}

func (h *BlockCacheImpl) getHashMap(hash []byte) (*BlockCacheTree, bool) {
	h.hmlock.RLock()
	defer h.hmlock.RUnlock()
	rtn, ok := h.hashMap[string(hash)]
	return rtn, ok
}

func (h *BlockCacheImpl) setHashMap(hash []byte, bct *BlockCacheTree) {
	h.hmlock.Lock()
	defer h.hmlock.Unlock()
	h.hashMap[string(hash)] = bct
}

// Add 把块加入缓存
// block 块, verifier 块的验证函数
func (h *BlockCacheImpl) Add(blk *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) error {
	var code CacheStatus
	var newTree *BlockCacheTree
	bct, ok := h.getHashMap(blk.HeadHash())
	if ok {
		code, newTree = Duplicate, nil
	} else {
		bct, ok = h.getHashMap(blk.Head.ParentHash)
		if ok {
			code, newTree = h.addSubTree(bct, newBct(blk, bct), verifier)
		} else {
			code, newTree = NotFound, nil
		}
	}

	switch code {
	case Extend:
		fallthrough
	case Fork:
		h.setHashMap(blk.HeadHash(), newTree)
		h.addSingles(newTree, verifier)
		if newTree.bctType == Singles {
			return ErrNotFound
		}
		h.tryFlush(blk.Head.Version)
	case NotFound:
		// Add to single block tree
		newTree = newBct(blk, h.singleBlockRoot)
		h.singleBlockRoot.children = append(h.singleBlockRoot.children, newTree)
		h.setHashMap(blk.HeadHash(), newTree)
		h.addSingles(newTree, verifier)
		return ErrNotFound
	case Duplicate:
		return ErrDup
	case ErrorBlock:
		return ErrBlock
	}
	return nil
}

// addSingles 尝试把single blocks上链
func (h *BlockCacheImpl) addSingles(newTree *BlockCacheTree, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) {
	block := newTree.bc.Top()
	newChildren := make([]*BlockCacheTree, 0)
	for k, _ := range h.singleBlockRoot.children {
		//fmt.Println(bct.bc.block.Head.ParentHash)
		if bytes.Equal(h.singleBlockRoot.children[k].bc.Top().Head.ParentHash, block.HeadHash()) {
			h.addSubTree(newTree, h.singleBlockRoot.children[k], verifier)
		} else {
			newChildren = append(newChildren, h.singleBlockRoot.children[k])
		}
	}
	h.singleBlockRoot.children = newChildren
}

func (h *BlockCacheImpl) addSubTree(root *BlockCacheTree, child *BlockCacheTree, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) (CacheStatus, *BlockCacheTree) {
	blk := child.bc.Top()
	newTree := newBct(blk, root)
	if root.bctType == OnCache {
		newPool, err := verifier(blk, root.bc.Top(), root.pool)
		if err != nil {
			h.hmlock.Lock()
			h.delSubTree(child)
			h.hmlock.Unlock()
			log.Log.I("verify block failed. err=%v", err)
			return ErrorBlock, nil
		}
		newTree.pool = newPool
	}
	newTree.bctType = root.bctType
	h.setHashMap(blk.HeadHash(), newTree)
	for k, _ := range child.children {
		h.addSubTree(newTree, child.children[k], verifier)
	}
	root.children = append(root.children, newTree)
	if len(root.children) == 1 {
		return Extend, newTree
	} else {
		return Fork, newTree
	}
}

func (h *BlockCacheImpl) delSubTree(root *BlockCacheTree) {
	delete(h.hashMap, string(root.bc.Top().HeadHash()))
	for _, bct := range root.children {
		h.delSubTree(bct)
	}
}

func (h *BlockCacheImpl) tryFlush(version int64) {
	for {
		// 可能进行多次flush
		need, newRoot := h.needFlush(version)
		if need {
			h.hmlock.Lock()
			for _, bct := range h.cachedRoot.children {
				if bct != newRoot {
					h.delSubTree(bct)
				}
			}
			delete(h.hashMap, string(h.cachedRoot.bc.Top().HeadHash()))
			h.hmlock.Unlock()
			h.cachedRoot = newRoot
			h.cachedRoot.bc.Flush()
			h.cachedRoot.pool.Flush()
			h.cachedRoot.super = nil
			h.cachedRoot.updateLength()

		} else {
			break
		}
	}
}

func (h *BlockCacheImpl) needFlush(version int64) (bool, *BlockCacheTree) {
	// TODO: 在底层parameter定义的地方定义各种version的const，可以在块生成、验证、此处用
	switch version {
	case 0:
		// PoB：确认某块的witness数大于maxDepth
		for _, bct := range h.cachedRoot.children {
			if bct.bc.confirmed > h.maxDepth {
				return true, bct
			}
		}
		return false, nil
	case 1:
		// PoW：最长链长度大于maxDepth
		if h.cachedRoot.bc.depth > h.maxDepth {
			return true, h.cachedRoot.popLongest()
		}
		return false, nil
	}
	return false, nil
}

// FindBlockInCache 在缓存中找一个块，根据块的hash
func (h *BlockCacheImpl) FindBlockInCache(hash []byte) (*block.Block, error) {
	var pb *block.Block
	found := h.cachedRoot.iterate(func(bct *BlockCacheTree) bool {
		if bytes.Equal(bct.bc.Top().HeadHash(), hash) {
			pb = bct.bc.Top()
			return true
		} else {
			return false
		}
	})

	if found {
		return pb, nil
	} else {
		return nil, fmt.Errorf("not found")
	}
}

func (h *BlockCacheImpl) CheckBlock(hash []byte) bool {
	_, ok := h.getHashMap(hash)
	if ok {
		return true
	}
	blk := h.bc.GetBlockByHash(hash)
	if blk != nil {
		return true
	}
	return false
}

// LongestChain 返回缓存的最长链
func (h *BlockCacheImpl) LongestChain() block.Chain {
	bct := h.cachedRoot
	for {
		if len(bct.children) == 0 {
			return &bct.bc
		}
		for _, b := range bct.children {
			if b.bc.depth == bct.bc.depth-1 {
				bct = b
				break
			}
		}
	}
}

func (h *BlockCacheImpl) BasePool() state.Pool {
	return h.cachedRoot.pool
}

func (h *BlockCacheImpl) SetBasePool(statePool state.Pool) error {
	h.cachedRoot.pool = statePool
	return nil
}

// LongestPool 返回最长链对应的state池
func (h *BlockCacheImpl) LongestPool() state.Pool {
	bct := h.cachedRoot
	for {
		if len(bct.children) == 0 {
			return bct.pool
		}
		for _, b := range bct.children {
			if b.bc.depth == bct.bc.depth-1 {
				bct = b
				break
			}
		}
	}
}

// BlockConfirmChan 返回块确认通道
func (h *BlockCacheImpl) BlockConfirmChan() chan uint64 {
	return h.blkConfirmChan
}

func (h *BlockCacheImpl) OnBlockChan() chan *block.Block {
	return h.chConfirmBlockData
}

func (h *BlockCacheImpl) SendOnBlock(blk *block.Block) {
	h.chConfirmBlockData <- blk
}

//func (h *BlockCacheImpl) FindTx(txHash []byte) (core.Tx, error) {
//	return core.Tx{}, nil
//}

//func (h *BlockCacheImpl) FindTxInCache(txHash []byte) (core.Tx, error) {
//	var tx core.Tx
//	var txp core.TxPoolImpl
//	var err error
//	for _, blk := range h.singleBlocks {
//		txp.Decode(blk.Content)
//		tx, err := txp.Find(txHash)
//		if err == nil {
//			return tx, err
//		}
//	}
//	found := h.cachedRoot.iterate(func(bct *BlockCacheTree) bool {
//		txp.Decode(bct.blk.Content)
//		tx, err = txp.Find(txHash)
//		if err == nil {
//			return true
//		} else {
//			return false
//		}
//	})
//
//	if found {
//		return tx, err
//	} else {
//		return tx, fmt.Errorf("not found")
//	}
//}

//func (h *BlockCacheImpl) LongestPool() core.UTXOPool {
//	bct := h.cachedRoot
//	for {
//		if bct.depth == 0 {
//			return bct.pool
//		}
//		for _, b := range bct.children {
//			if b.depth == bct.depth-1 {
//				bct = b
//				break
//			}
//		}
//	}
//}

//for debug
//draw the blockcache
const PICSIZE int = 100

var pic [PICSIZE][PICSIZE]byte
var picX, picY int

func calcTree(root *BlockCacheTree, x int, y int, isLast bool) int {
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
	for i := 0; i < len(root.children); i++ {
		if i == len(root.children)-1 {
			f = true
		}
		width = calcTree(root.children[i], x+width, y+2, f)
	}
	if isLast {
		return x + width
	} else {
		return x + width + 2
	}
}
func (b *BlockCacheTree) DrawTree() {
	for i := 0; i < PICSIZE; i++ {
		for j := 0; j < PICSIZE; j++ {
			pic[i][j] = ' '
		}
	}
	calcTree(b, 0, 0, true)
	for i := 0; i <= picX; i++ {
		for j := 0; j <= picY; j++ {
			fmt.Printf("%c", pic[i][j])
		}
		fmt.Printf("\n")
	}
}

func (h *BlockCacheImpl) Draw() {
	fmt.Println("\ncachedTree:")
	h.cachedRoot.DrawTree()
	//	fmt.Println("cachedSingle:")
	//	h.singleBlockRoot.DrawTree()
}
