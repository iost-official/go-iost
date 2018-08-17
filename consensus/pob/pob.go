package pob

import (
	"errors"
	"fmt"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/db"
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
	synchronizer synchronizer.Synchronizer
	verifyDB     db.MVCCDB
	produceDB    db.MVCCDB

	exitSignal  chan struct{}
	chRecvBlock chan p2p.IncomingMessage
	chGenBlock  chan *block.Block
}

func NewPoB(account account.Account, baseVariable global.BaseVariable, blockCache blockcache.BlockCache, txPool txpool.TxPool, p2pService p2p.Service, synchronizer synchronizer.Synchronizer, witnessList []string) *PoB {
	p := PoB{
		account:      account,
		baseVariable: baseVariable,
		blockChain:   baseVariable.BlockChain(),
		blockCache:   blockCache,
		txPool:       txPool,
		p2pService:   p2pService,
		synchronizer: synchronizer,
		verifyDB:     baseVariable.StateDB(),
		produceDB:    baseVariable.StateDB().Fork(),
		exitSignal:   make(chan struct{}),
		chRecvBlock:  p2pService.Register("consensus channel", p2p.NewBlockResponse, p2p.SyncBlockResponse),
		chGenBlock:   make(chan *block.Block, 10),
	}
	staticProperty = newStaticProperty(p.account, witnessList)
	return &p
}

func (p *PoB) Run() {
	p.synchronizer.Start()
	go p.blockLoop()
	go p.scheduleLoop()
}

func (p *PoB) Stop() {
	p.synchronizer.Stop()
	close(p.exitSignal)
	close(p.chRecvBlock)
	close(p.chGenBlock)
}

func (p *PoB) blockLoop() {
	for {
		select {
		case incomingMessage, ok := <-p.chRecvBlock:
			if !ok {
				fmt.Println("chRecvBlock has closed")
				return
			}
			var blk block.Block
			err := blk.Decode(incomingMessage.Data())
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = p.handleRecvBlock(&blk)
			if err != nil {
				fmt.Println(err)
				continue
			}
			go p.p2pService.Broadcast(req.Data(), req.Type(), p2p.UrgentMessage)
			if incomingMessage.Type() == p2p.SyncBlockResponse {
				go p.synchronizer.OnBlockConfirmed(string(blk.HeadHash()), req.From())
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
			if witnessOfSec(time.Now().Unix()) == p.account.ID && p.baseVariable.Mode().Mode() == global.ModeNormal {
				blk, err := generateBlock(p.account, p.blockCache.Head().Block, p.txPool, p.produceDB)
				if err != nil {
					fmt.Println(err)
					fmt.Println("fail to generateBlock")
					continue
				}
				blkByte, err := blk.Encode()
				if err != nil {
					fmt.Println(err)
					continue
				}
				p.chGenBlock <- blk
				go p.p2pService.Broadcast(blkByte, p2p.NewBlockResponse, p2p.UrgentMessage)
			}
			nextSchedule = timeUntilNextSchedule(time.Now().Unix())
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) handleRecvBlock(blk *block.Block) error {
	_, err := p.blockCache.Find(blk.HeadHash())
	if err == nil {
		return errors.New("duplicate block")
	}
	err = verifyBasics(blk)
	if err != nil {
		return errors.New("fail to verifyBasics")
	}
	parent, err := p.blockCache.Find(blk.Head.ParentHash)
	if err == nil && parent.Type == blockcache.Linked {
		return p.addNewBlock(blk, parent.Block) // only need to consider error from addNewBlock, not from addExistingblock
	} else {
		p.blockCache.Add(blk)
	}
	return nil
}

func (p *PoB) addNewBlock(blk *block.Block, parentBlock *block.Block) error {
	if blk.Head.Witness != p.account.ID {
		p.verifyDB.Checkout(string(blk.Head.ParentHash))
		err := verifyBlock(blk, parentBlock, p.blockCache.LinkedRoot().Block, p.txPool, p.verifyDB)
		if err != nil {
			return err
		}
		p.verifyDB.Tag(string(blk.HeadHash()))
	} else {
		p.verifyDB.Checkout(string(blk.HeadHash()))
	}
	node := p.blockCache.Add(blk)
	p.updateInfo(node)
	p.addChildren(node)
	return nil
}

func (p *PoB) addExistingBlock(blk *block.Block, parentBlock *block.Block) {
	node, _ := p.blockCache.Find(blk.HeadHash())
	if blk.Head.Witness != p.account.ID {
		p.verifyDB.Checkout(string(blk.Head.ParentHash))
		err := verifyBlock(blk, parentBlock, p.blockCache.LinkedRoot().Block, p.txPool, p.verifyDB)
		if err != nil {
			p.blockCache.Del(node)
			fmt.Println(err)
		}
		p.verifyDB.Tag(string(blk.HeadHash()))
	} else {
		p.verifyDB.Checkout(string(blk.HeadHash()))
	}
	p.blockCache.Link(node)
	p.updateInfo(node)
	p.addChildren(node)
}

func (p *PoB) addChildren(node *blockcache.BlockCacheNode) {
	for child := range node.Children {
		p.addExistingBlock(child.Block, node.Block)
	}
}

func (p *PoB) updateInfo(node *blockcache.BlockCacheNode) {
	updateStaticProperty(node)
	updatePendingWitness(node, p.verifyDB)
	updateLib(node, p.blockCache)
	p.txPool.AddLinkedNode(node, p.blockCache.Head())
}
