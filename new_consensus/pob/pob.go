package pob

import (
	"errors"
	"fmt"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/log"
	"github.com/iost-official/Go-IOS-Protocol/new_consensus/common"
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
	account      account.Account
	baseVariable global.BaseVariable
	blockChain   block.Chain
	blockCache   blockcache.BlockCache
	txPool       txpool.TxPool
	p2pService   p2p.Service
	synchronizer consensus_common.Synchronizer
	verifyDB     *db.MVCCDB
	produceDB    *db.MVCCDB

	exitSignal  chan struct{}
	chRecvBlock chan message.Message
	chGenBlock  chan *block.Block
}

func NewPoB(account account.Account, baseVariable global.BaseVariable, blockCache blockcache.BlockCache, txPool txpool.TxPool, p2pService p2p.Service, synchronizer consensus_common.Synchronizer, witnessList []string) (*PoB, error) {
	//TODO: change initialization based on new interfaces
	p := PoB{
		account:      account,
		baseVariable: baseVariable,
		blockCache:   blockCache,
		blockChain:   baseVariable.BlockChain(),
		verifyDB:     baseVariable.StateDB(),
		txPool:       txPool,
		p2pService:   p2pService,
		synchronizer: synchronizer,
		chGenBlock:   make(chan *block.Block, 10),
	}

	p.produceDB = p.verifyDB.Fork()

	var err error
	//p.chRecvBlock, err = p.p2pService.Register("consensus chan", p2p.NewBlockResponse, p2p.SyncBlockResponse)
	if err != nil {
		return nil, err
	}
	p.exitSignal = make(chan struct{})
	p.initGlobalProperty(p.account, witnessList)
	blk, err := p.blockChain.Top()
	if err != nil {
		fmt.Println("Unable to initialize block chain top")
	}
	dynamicProperty.update(blk)
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
				//go p.synchronizer.OnRecvBlock(blk.HeadHash(), req.From())
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
			currentTimestamp := common.GetCurrentTimestamp()
			wid := witnessOfTime(currentTimestamp)
			if wid == p.account.ID && p.baseVariable.Mode().Mode() == global.ModeNormal {
				chainHead := p.blockCache.Head()
				hash := chainHead.Block.HeadHash()
				p.produceDB.Checkout(string(hash))
				blk := genBlock(p.account, chainHead, p.txPool, p.produceDB)
				dynamicProperty.update(blk)
				blkByte, err := blk.Encode()
				if err != nil {
					fmt.Println(err)
				}
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
	if blk.Head.Witness != p.account.ID {
		hash := parent.Block.HeadHash()
		p.verifyDB.Checkout(string(hash))
		var verifyErr error
		verifyErr = verifyBlock(blk, parent.Block, p.blockCache.LinkedRoot().Block, p.txPool, p.verifyDB)
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
		p.verifyDB.Tag(string(blk.HeadHash()))
	} else {
		p.verifyDB.Checkout(string(blk.HeadHash()))
	}
	updateNodeInfo(node)
	updatePendingWitness(node, p.verifyDB) //?
	confirmNode := calculateConfirm(node, p.blockCache.LinkedRoot())
	p.blockCache.Flush(confirmNode)
	promoteWitness(node, confirmNode)	//更新witnesslist?
	dynamicProperty.update(blk)
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
