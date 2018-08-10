package pob

import (
	. "github.com/iost-official/Go-IOS-Protocol/account"
	. "github.com/iost-official/Go-IOS-Protocol/new_consensus/common"

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
	account      Account
	global       global.Global
	blockChain   block.Chain
	blockCache   blockcache.BlockCache
	txPool       new_txpool.TxPool
	p2pService   p2p.Service
	synchronizer *Synchronizer
	verifyDB     *db.MVCCDB
	produceDB    *db.MVCCDB

	exitSignal  chan struct{}
	chRecvBlock chan message.Message
	chGenBlock  chan *block.Block

	log *log.Logger
}

func NewPoB(acc Account, global global.Global, blkcache blockcache.BlockCache, p2pserv p2p.Service, sy *Synchronizer, witnessList []string) (*PoB, error) {
	//TODO: change initialization based on new interfaces
	p := PoB{
		account:      acc,
		global:       global,
		blockCache:   blkcache,
		blockChain:   global.BlockChain(),
		verifyDB:     global.StdPool(),
		txPool:       global.TxPool(),
		p2pService:   p2pserv,
		synchronizer: sy,
		chGenBlock:   make(chan *block.Block, 10),
	}

	p.produceDB = p.verifyDB.Fork()
	/*
		if p.blockChain.GetBlockByNumber(0) == nil {

			t := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
			genesis := genGenesis(GetTimestamp(t.Unix()).Slot)
			//TODO: add genesis to db, what about its state?
			p.blockChain.Push(genesis)
		}
	*/

	var err error
	p.chRecvBlock, err = p.p2pService.Register("consensus chan", p2p.NewBlockResponse, p2p.SyncBlockResponse)
	if err != nil {
		return nil, err
	}
	p.exitSignal = make(chan struct{})

	p.log, err = log.NewLogger("consensus.log")
	if err != nil {
		return nil, err
	}

	p.log.NeedPrint = false

	p.initGlobalProperty(p.account, witnessList)

	if blk, err := p.blockChain.Top(); err == nil {
		dynamicProp.update(&blk.Head)
	} else {
		p.log.E("Unable to initialize block chain top")
	}
	return &p, nil
}

func (p *PoB) initGlobalProperty(acc Account, witnessList []string) {
	staticProp = newGlobalStaticProperty(acc, witnessList)
	dynamicProp = newGlobalDynamicProperty()
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

func (p *PoB) handleRecvBlock(blk *block.Block) bool {
	hash, err := blk.HeadHash()
	if err != nil {
		return false
	}
	if _, err := p.blockCache.Find(hash); err == nil {
		p.log.I("Duplicate block: %v", hash)
		return false
	}
	if err := verifyBasics(blk); err == nil {
		parent, err := p.blockCache.Find(blk.Head.ParentHash)
		if err == nil && parent.Type == blockcache.Linked {
			// Can be linked
			// tell synchronizer to cancel downloading

			var node *blockcache.BlockCacheNode
			var err error
			node, err = p.addBlock(blk, node, parent, true)
			if err != nil {
				// dishonest?
				p.log.I("Add block error: %v", err)
				return false
			}
			p.addSingles(node)
		} else {
			// Single block
			p.blockCache.Add(blk)
		}
	} else {
		// dishonest?
		p.log.I("Add block error: %v", err)
		return false
	}
	return true
}

func (p *PoB) blockLoop() {
	p.log.I("Start to listen block")
	for {
		select {
		case req, ok := <-p.chRecvBlock:
			if !ok {
				return
			}
			var blk block.Block
			err := blk.Decode(req.GetBody())
			if err != nil {
				continue
			}
			if p.handleRecvBlock(&blk) {
				if req.Type() == p2p.SyncBlockResponse {
					hash, err := blk.HeadHash()
					if err == nil {
						go p.synchronizer.OnRecvBlock(hash, req.From())
					}
				}
			}
		case blk, ok := <-p.chGenBlock:
			if !ok {
				return
			}
			p.handleRecvBlock(blk)
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) scheduleLoop() {
	var nextSchedule int64
	nextSchedule = 0
	p.log.I("Start to schedule")
	for {
		select {
		case <-time.After(time.Second * time.Duration(nextSchedule)):
			currentTimestamp := GetCurrentTimestamp()
			wid := witnessOfTime(currentTimestamp)
			p.log.I("currentTimestamp: %v, wid: %v, p.account.ID: %v", currentTimestamp, wid, p.account.ID)
			if wid == p.account.ID && p.global.Mode() == global.ModeNormal {
				chainHead := p.blockCache.Head()
				hash, err := chainHead.Block.HeadHash()
				if err != nil {
					continue
				}
				p.produceDB.Checkout(string(hash))
				blk := genBlock(p.account, chainHead, p.txPool, p.produceDB)

				dynamicProp.update(&blk.Head)
				p.log.I("Generating block, current timestamp: %v number: %v", currentTimestamp, blk.Head.Number)

				bb := blk.Encode()
				//msg := message.Message{ReqType: int32(ReqNewBlock), Body: bb}
				//go p.router.Broadcast(msg)
				p.chGenBlock <- blk
				log.Log.I("Block size: %v, TrNum: %v", len(bb), len(blk.Txs))
				go p.p2pService.Broadcast(bb, p2p.NewBlockResponse, p2p.UrgentMessage)
				p.log.I("Broadcasted block, current timestamp: %v number: %v", currentTimestamp, blk.Head.Number)
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
		hash, err := parent.Block.HeadHash()
		if err != nil {
			return nil, err
		}
		p.verifyDB.Checkout(string(hash))
		var verifyErr error
		verifyErr = verifyBlock(blk, parent.Block, p.blockCache.LinkedRoot().Block, p.txPool, p.verifyDB)

		// add
		if newBlock {
			if verifyErr == nil {
				var err error
				if node, err = p.blockCache.Add(blk); err != nil {
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
		hash, err = blk.HeadHash()
		if err != nil {
			return nil, err
		}
		p.verifyDB.Tag(string(hash))
	} else {
		hash, err := blk.HeadHash()
		if err != nil {
			return nil, err
		}
		p.verifyDB.Checkout(string(hash))
	}

	// update node info without state
	updateNodeInfo(node)
	// update node info with state, currently pending witness list
	updatePendingWitness(node, p.verifyDB)

	// confirm
	confirmNode := calculateConfirm(node, p.blockCache.LinkedRoot())
	if confirmNode != nil {
		p.blockCache.Flush(confirmNode)
		// promote witness list
		promoteWitness(node, confirmNode)
	}

	dynamicProp.update(&blk.Head)
	// -> tx pool
	p.txPool.AddLinkedNode(node, p.blockCache.Head())
	return node, nil
}

func (p *PoB) addSingles(node *blockcache.BlockCacheNode) {
	if node.Children != nil {
		for child := range node.Children {
			if _, err := p.addBlock(nil, child, node, false); err == nil {
				p.addSingles(child)
			}
		}
	}
}
