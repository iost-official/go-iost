package pob

import (
	"errors"
	"fmt"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
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

var errSingle = errors.New("single blcok")

func init() {
	prometheus.MustRegister(generatedBlockCount)
	prometheus.MustRegister(receivedBlockCount)
	prometheus.MustRegister(confirmedBlockchainLength)
	prometheus.MustRegister(txPoolSize)
}

//PoB is a struct that handles the consensus logic.
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
	exitSignal   chan struct{}
	chRecvBlock  chan p2p.IncomingMessage
	chGenBlock   chan *block.Block
}

// NewPoB init a new PoB.
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
		chRecvBlock:  p2pService.Register("consensus channel", p2p.NewBlock, p2p.SyncBlockResponse),
		chGenBlock:   make(chan *block.Block, 10),
	}
	staticProperty = newStaticProperty(p.account, witnessList)
	return &p
}

//Start make the PoB run.
func (p *PoB) Start() error {
	go p.blockLoop()
	go p.scheduleLoop()
	return nil
}

//Stop make the PoB stop.
func (p *PoB) Stop() {
	close(p.exitSignal)
	close(p.chRecvBlock)
	close(p.chGenBlock)
}

func (p *PoB) blockLoop() {
	ilog.Infof("start block")
	for {
		select {
		case incomingMessage, ok := <-p.chRecvBlock:
			if !ok {
				ilog.Infof("chRecvBlock has closed")
				return
			}
			var blk block.Block
			err := blk.Decode(incomingMessage.Data())
			if err != nil {
				ilog.Error(err.Error())
				continue
			}
			ilog.Info("new block come, block number: ", blk.Head.Number)
			err = p.handleRecvBlock(&blk)
			if err != nil && err != errSingle {
				ilog.Error(err.Error())
				continue
			}
			p.synchronizer.CheckSyncProcess()
			if incomingMessage.Type() == p2p.SyncBlockResponse {
				go p.synchronizer.OnBlockConfirmed(string(blk.HeadHash()), incomingMessage.From())
			}
			if incomingMessage.Type() == p2p.NewBlock {
				go p.p2pService.Broadcast(incomingMessage.Data(), incomingMessage.Type(), p2p.UrgentMessage)
				ilog.Info("err type ", err)
				if need, start, end := p.synchronizer.NeedSync(blk.Head.Number); need && (err == errSingle) {
					go p.synchronizer.SyncBlocks(start, end)
				}
			}
			p.blockCache.Draw()
		case blk, ok := <-p.chGenBlock:
			if !ok {
				ilog.Infof("chGenBlock has closed")
				return
			}
			ilog.Info("block from myself, block number: ", blk.Head.Number)
			err := p.handleRecvBlock(blk)
			if err != nil {
				ilog.Error(err.Error())
			}
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) scheduleLoop() {
	nextSchedule := timeUntilNextSchedule(time.Now().UnixNano())
	ilog.Infof("nextSchedule: %.2f", time.Duration(nextSchedule).Seconds())
	for {
		select {
		case <-time.After(time.Duration(nextSchedule)):
			ilog.Infof("nextSchedule: %.2f", time.Duration(nextSchedule).Seconds())
			ilog.Info(p.baseVariable.Mode().Mode())
			if witnessOfSec(time.Now().Unix()) == p.account.ID {
				ilog.Info(p.baseVariable.Mode().Mode())
				if p.baseVariable.Mode().Mode() == global.ModeNormal {
					blk, err := generateBlock(p.account, p.blockCache.Head().Block, p.txPool, p.produceDB)
					ilog.Infof("gen block:%v", blk.Head.Number)
					if err != nil {
						ilog.Error(err.Error())
						continue
					}
					blkByte, err := blk.Encode()
					if err != nil {
						ilog.Error(err.Error())
						continue
					}
					p.chGenBlock <- blk
					go p.p2pService.Broadcast(blkByte, p2p.NewBlock, p2p.UrgentMessage)
				}
				time.Sleep(common.SlotLength * time.Second)
			}
			nextSchedule = timeUntilNextSchedule(time.Now().UnixNano())
			ilog.Infof("nextSchedule: %.2f", time.Duration(nextSchedule).Seconds())
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
		return fmt.Errorf("fail to verify blocks, %v", err)
	}
	parent, err := p.blockCache.Find(blk.Head.ParentHash)
	p.blockCache.Add(blk)
	staticProperty.addSlot(blk.Head.Time)
	if err == nil && parent.Type == blockcache.Linked {
		return p.addExistingBlock(blk, parent.Block)
	} else {
		return errSingle
	}
}

func (p *PoB) addExistingBlock(blk *block.Block, parentBlock *block.Block) error {
	node, _ := p.blockCache.Find(blk.HeadHash())
	if blk.Head.Witness != p.account.ID {
		p.verifyDB.Checkout(string(blk.Head.ParentHash))
		err := verifyBlock(blk, parentBlock, p.blockCache.LinkedRoot().Block, p.txPool, p.verifyDB)
		if err != nil {
			p.blockCache.Del(node)
			ilog.Error(err.Error())
			return err
		}
		p.verifyDB.Tag(string(blk.HeadHash()))
	} else {
		p.verifyDB.Checkout(string(blk.HeadHash()))
	}
	p.blockCache.Link(node)
	p.updateInfo(node)
	for child := range node.Children {
		p.addExistingBlock(child.Block, node.Block)
	}
	return nil
}

func (p *PoB) updateInfo(node *blockcache.BlockCacheNode) {
	updateWaterMark(node)
	updateLib(node, p.blockCache)
	p.txPool.AddLinkedNode(node, p.blockCache.Head())
}
