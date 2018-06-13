package consensus_common

import (
	"bytes"
	"fmt"

	"errors"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/tx"
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

// BlockCacheTree 缓存链分叉的树结构
type BlockCacheTree struct {
	bc       CachedBlockChain
	children []*BlockCacheTree
	super    *BlockCacheTree
	pool     state.Pool
}

func newBct(block *block.Block, tree *BlockCacheTree) *BlockCacheTree {
	bct := BlockCacheTree{
		bc:       tree.bc.Copy(),
		children: make([]*BlockCacheTree, 0),
		super:    tree,
	}

	bct.bc.Push(block)
	return &bct
}

func (b *BlockCacheTree) add(block *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) (CacheStatus, *BlockCacheTree) {
	for _, bct := range b.children {
		code, newTree := bct.add(block, verifier)
		if code != NotFound {
			return code, newTree
		}
	}

	if bytes.Equal(b.bc.Top().Head.Hash(), block.Head.ParentHash) {
		for _, bct := range b.children {
			if bytes.Equal(bct.bc.Top().Head.Hash(), block.Head.Hash()) {
				//fmt.Println(block.Head.Hash())
				return Duplicate, nil
			}
		}
		newPool, err := verifier(block, b.bc.Top(), b.pool)
		if err != nil {
			fmt.Printf("ErrorBlock: %v\n", err)

			return ErrorBlock, nil
		}

		bct := newBct(block, b)
		bct.pool = newPool
		b.children = append(b.children, bct)
		if len(b.children) == 1 {
			return Extend, bct
		} else {
			return Fork, bct
		}
	}

	return NotFound, nil
}

func (b *BlockCacheTree) findSingles(block *block.Block) (error, *BlockCacheTree) {
	if b.bc.block != nil && bytes.Equal(b.bc.block.Head.Hash(), block.Head.Hash()) {
		return ErrDup, nil
	}
	for _, bct := range b.children {
		err, ret := bct.findSingles(block)
		if err == nil {
			return err, ret
		} else if err == ErrDup {
			return err, nil
		}
	}
	if b.bc.block != nil && bytes.Equal(b.bc.block.Head.Hash(), block.Head.ParentHash) {
		return nil, b
	}
	return ErrNotFound, nil
}

func (b *BlockCacheTree) addSubTree(root *BlockCacheTree, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) {
	block := root.bc.block
	newPool, err := verifier(block, b.bc.Top(), b.pool)
	if err != nil {
		return
	}

	newTree := newBct(block, b)
	newTree.pool = newPool
	b.children = append(b.children, newTree)
	for _, bct := range root.children {
		newTree.addSubTree(bct, verifier)
	}
}

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

// BlockCache 操作块缓存的接口
type BlockCache interface {
	AddGenesis(block *block.Block) error
	Add(block *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) error
	AddSingles(verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error))
	AddTx(tx *tx.Tx) error
	GetTx() (*tx.Tx, error)
	UpdateTxPoolOnBC(bc block.Chain)
	ResetTxPoool() error

	FindBlockInCache(hash []byte) (*block.Block, error)
	LongestChain() block.Chain
	LongestPool() state.Pool
	BlockChain() block.Chain
	BasePool() state.Pool
	SetBasePool(statePool state.Pool) error
	ConfirmedLength() uint64
	BlockConfirmChan() chan uint64
	Draw()
}

// BlockCacheImpl 块缓存实现
type BlockCacheImpl struct {
	bc              block.Chain
	cachedRoot      *BlockCacheTree
	singleBlockRoot *BlockCacheTree
	recentTree      *BlockCacheTree
	txPool          tx.TxPool
	delTxPool       tx.TxPool
	txPoolCache     tx.TxPool
	maxDepth        int
	blkConfirmChan  chan uint64
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
		},
		singleBlockRoot: &BlockCacheTree{
			bc: CachedBlockChain{
				block: nil,
			},
			children: make([]*BlockCacheTree, 0),
			super:    nil,
		},
		maxDepth:       maxDepth,
		blkConfirmChan: make(chan uint64, 10),
	}
	h.txPool, _ = tx.TxPoolFactory("mem")
	h.txPoolCache, _ = tx.TxPoolFactory("mem")
	h.delTxPool, _ = tx.TxPoolFactory("mem")
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
	h.bc.Push(block)
	h.cachedRoot.pool.Flush()
	return nil
}

// Add 把块加入缓存
// block 块, verifier 块的验证函数
func (h *BlockCacheImpl) Add(block *block.Block, verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) error {
	/*
		if uint64(block.Head.Number) < h.bc.Length() {
			return ErrTooOld
		}
	*/
	code, newTree := h.cachedRoot.add(block, verifier)
	h.recentTree = newTree
	switch code {
	case Extend:
		fallthrough
	case Fork:
		h.tryFlush(block.Head.Version)
	case NotFound:
		// Add to single block tree
		err, bct := h.singleBlockRoot.findSingles(block)
		if err == ErrNotFound {
			bct = h.singleBlockRoot
		} else if err == ErrDup {
			return ErrDup
		}
		newTree := &BlockCacheTree{
			bc: CachedBlockChain{
				block: block,
			},
			super:    bct,
			children: make([]*BlockCacheTree, 0),
		}

		bct.children = append(bct.children, newTree)

		newChildren := make([]*BlockCacheTree, 0)
		for _, bct := range h.singleBlockRoot.children {
			if bytes.Equal(bct.bc.block.Head.ParentHash, block.Head.Hash()) {
				newTree.children = append(newTree.children, bct)
				bct.super = newTree
			} else {
				newChildren = append(newChildren, bct)
			}
		}
		h.singleBlockRoot.children = newChildren
		return ErrNotFound
	case Duplicate:
		return ErrDup
	case ErrorBlock:
		return ErrBlock
	}
	return nil
}

// AddSingles 尝试把single blocks上链
func (h *BlockCacheImpl) AddSingles(verifier func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error)) {
	newTree := h.recentTree
	block := newTree.bc.block
	newChildren := make([]*BlockCacheTree, 0)
	for _, bct := range h.singleBlockRoot.children {
		fmt.Println(bct.bc.block.Head.ParentHash)
		if bytes.Equal(bct.bc.block.Head.ParentHash, block.Head.Hash()) {
			newTree.addSubTree(bct, verifier)
		} else {
			newChildren = append(newChildren, bct)
		}
	}
	h.singleBlockRoot.children = newChildren
	h.tryFlush(block.Head.Version)
}

// AddTx 把交易加入链
func (h *BlockCacheImpl) AddTx(tx *tx.Tx) error {
	//TODO 验证tx是否在blockchain上
	if ok, _ := h.bc.HasTx(tx); ok {
		return fmt.Errorf("tx in BlockChain")
	}
	h.txPool.Add(tx)
	return nil
}

// GetTx 从链中取交易
func (h *BlockCacheImpl) GetTx() (*tx.Tx, error) {
	for {
		tx, err := h.txPool.Top()
		if err != nil {
			return nil, err
		}
		h.txPool.Del(tx)
		h.txPoolCache.Add(tx)
		if ok, _ := h.delTxPool.Has(tx); !ok {
			return tx, nil
		}
	}
}

func (h *BlockCacheImpl) UpdateTxPoolOnBC(bc block.Chain) {
	h.delTxPool = tx.NewTxPoolImpl()
	iter := bc.Iterator()
	for {
		block := iter.Next()
		if block == nil {
			break
		}
		for _, tx := range block.Content {
			h.delTxPool.Add(&tx)
		}
	}
}

func (h *BlockCacheImpl) ResetTxPoool() error {
	for h.txPoolCache.Size() > 0 {
		tx, _ := h.txPoolCache.Top()
		h.AddTx(tx)
		h.txPoolCache.Del(tx)
	}
	return nil
}

func (h *BlockCacheImpl) tryFlush(version int64) {
	for {
		// 可能进行多次flush
		need, newRoot := h.needFlush(version)
		if need {
			h.cachedRoot = newRoot
			h.cachedRoot.bc.Flush()
			h.cachedRoot.pool.Flush()
			h.cachedRoot.super = nil
			h.cachedRoot.updateLength()
			blockCachedLength.Set(float64(h.cachedRoot.bc.Length()))

			for _, tx := range h.cachedRoot.bc.Top().Content {
				h.txPool.Del(&tx)
			}
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
