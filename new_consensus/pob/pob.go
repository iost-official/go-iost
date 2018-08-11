package pob

import (
	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/new_consensus/common"

	"time"

	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/log"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/prometheus/client_golang/prometheus"
	"fmt"
	"errors"
	blockcache2 "github.com/iost-official/Go-IOS-Protocol/core/blockcache"
)

var (
	generatedBlockCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "generated_block_count",
			Help: "Count of generated block by current node",
		},
	)
	receivedBlockCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "received_block_count",
			Help: "Count of received block by current node",
		},
	)
	confirmedBlockchainLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "confirmed_blockchain_length",
			Help: "Length of confirmed blockchain on current node",
		},
	)
	txPoolSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "tx_poo_size",
			Help: "size of tx pool on current node",
		},
	)
)

func init() {
	prometheus.MustRegister(generatedBlockCount)
	prometheus.MustRegister(receivedBlockCount)
	prometheus.MustRegister(confirmedBlockchainLength)
	prometheus.MustRegister(txPoolSize)
}

type PoB struct {
	account      account.Account
	global       global.Global
	blockChain   block.Chain
	blockCache   blockcache.BlockCache
	txPool       new_txpool.TxPool
	p2pService   p2p.Service
	synchronizer consensus_common.Synchronizer
	addBlockPointer     *db.MVCCDB
	genBlockPointer    *db.MVCCDB

	exitSignal  chan struct{}
	chRecvBlock chan message.Message
	chGenBlock  chan *block.Block
}

func NewPoB(account_ account.Account, blockchain_ block.BlockChain, blockcache_ blockcache.BlockCache, txPool_ new_txpool.TxPool, service_ p2p.Service, synchronizer_ consensus_common.Synchronizer, witnessList []string) (*PoB, error) {
	//TODO: change initialization based on new interfaces
	p := PoB{
		account:      account_,
		global:       global_,
		blockCache:   blockcache_,
		blockChain:   global_.BlockChain(),
		addBlockPointer:     global_.StatePool(),
		txPool:       txPool_,
		p2pService:   service_,
		synchronizer: synchronizer_,
		chGenBlock:   make(chan *block.Block, 10),
	}

	p.genBlockPointer = p.addBlockPointer.Fork()

	var err error
	p.chRecvBlock, err = p.p2pService.Register("consensus chan", p2p.NewBlockResponse, p2p.SyncBlockResponse)
	if err != nil {
		return nil, err
	}
	p.exitSignal = make(chan struct{})
	p.initGlobalProperty(p.account, witnessList)
	blk, err := p.blockChain.Top()
	if err != nil {
		fmt.Println("Unable to initialize block chain top")
	}
	dynamicProperty.update(&blk.Head)
	return &p, nil
}

func (p *PoB) initGlobalProperty(acc account.Account, witnessList []string) {
	staticProperty = newGlobalStaticProperty(acc, witnessList)
	dynamicProperty = newGlobalDynamicProperty()
}

func (p *PoB) Run() {
	p.synchronizer.Start()
	go p.blockLoop()
	go p.scheduleLoop()
}

func (p *PoB) Stop() {
	p.synchronizer.Stop()
	close(p.chRecvBlock)
	close(p.chGenBlock)
	close(p.exitSignal)
}

func (p *PoB) handleRecvBlock(blk *block.Block) error {
	hash := blk.HeadHash()
	_, err := p.blockCache.Find(hash)
	if err == nil {
		return errors.New("duplicate block")
	}
	err = verifyBasics(blk)
	if err != nil {
		return errors.New("fail to verifyBasics")
	}
	parent, err := p.blockCache.Find(blk.Head.ParentHash)
	if err == nil && parent.Type == blockcache.Linked {
		var node *blockcache.BlockCacheNode
		node, err = p.addBlock(blk, node, parent, true)
		if err != nil {
			// dishonest?
			return errors.New("fail to addBlock")
		}
		p.addSingles(node)
	} else {
		p.blockCache.Add(blk)
	}
	return nil
}

func (p *PoB) blockLoop() {
	for {
		select {
		case req, ok := <-p.chRecvBlock:
			if !ok {
				fmt.Println("chRecvBlock has closed")
				return
			}
			var blk block.Block
			err := blk.Decode(req.GetBody())
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = p.handleRecvBlock(&blk)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if req.GetReqType() == int32(p2p.SyncBlockResponse) {
				go p.synchronizer.OnRecvBlock(blk.HeadHash(), req.From())
			}
		case blk, ok := <-p.chGenBlock:
			if !ok {
				fmt.Println("chGenBlock has closed")
				return
			}
			p.handleRecvBlock(blk)
		case <-p.exitSignal:
			fmt.Println("exitSignal")
			return
		}
	}
}

func (p *PoB) scheduleLoop() {
	var nextSchedule int64 = 0
	for {
		select {
		case <-time.After(time.Second * time.Duration(nextSchedule)):
			currentTimestamp := consensus_common.GetCurrentTimestamp()
			wid := witnessOfTime(currentTimestamp)
			if wid == p.account.ID && p.global.Mode() == global.ModeNormal {
				chainHead := p.blockCache.Head()
				hash, err := chainHead.Block.HeadHash()
				if err != nil {
					fmt.Println(err)
					continue
				}
				p.genBlockPointer.Checkout(string(hash))
				blk := genBlock(p.account, chainHead, p.txPool, p.genBlockPointer)
				dynamicProperty.update(&blk.Head)
				blkByte, err := blk.Encode()
				//msg := message.Message{ReqType: int32(ReqNewBlock), Body: bb}
				//go p.router.Broadcast(msg)
				p.chGenBlock <- blk
				log.Log.I("Block size: %v, TrNum: %v", len(blkByte), len(blk.Txs))
				go p.p2pService.Broadcast(blkByte, p2p.NewBlockResponse, p2p.UrgentMessage)
			}
			nextSchedule = timeUntilNextSchedule(time.Now().Unix())
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) addBlock(blk *block.Block, node *blockcache.BlockCacheNode, parent *blockcache.BlockCacheNode, newBlock bool) (*blockcache.BlockCacheNode, error) {
	// verify block txs
	if blk.Head.Witness != p.account.ID {
		hash := parent.Block.HeadHash()
		p.addBlockPointer.Checkout(string(hash))
		var verifyErr error
		verifyErr = verifyBlock(blk, parent.Block, p.blockCache.LinkedRoot().Block, p.txPool, p.addBlockPointer)

		// add
		if newBlock {
			if verifyErr == nil {
				var err error
				node, err = p.blockCache.Add(blk)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, verifyErr
			}
		} else {
			if verifyErr == nil {
				p.blockCache.Link(node)
			} else {
				p.blockCache.Del(node)
				return nil, verifyErr
			}
		}
		// tag in state
		hash = blk.HeadHash()
		p.addBlockPointer.Tag(string(hash))
	} else {
		hash := blk.HeadHash()
		p.addBlockPointer.Checkout(string(hash))
	}

	// update node info without state
	updateNodeInfo(node)
	// update node info with state, currently pending witness list
	updatePendingWitness(node, p.addBlockPointer)

	// confirm
	confirmNode := calculateConfirm(node, p.blockCache.LinkedRoot())
	if confirmNode != nil {
		p.blockCache.Flush(confirmNode)
		// promote witness list
		promoteWitness(node, confirmNode)
	}

	dynamicProperty.update(&blk.Head)
	// -> tx pool
	p.txPool.AddLinkedNode(node, p.blockCache.Head())
	return node, nil
}

func (p *PoB) addSingles(node *blockcache.BlockCacheNode) {
	for child := range node.Children {
		_, err := p.addBlock(child.Block, child, node, false)
		if err != nil {
			fmt.Println(err)
			continue
		}
		p.addSingles(child)
	}
}
