package pob

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/verifier"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/consensus/synchronizer/pb"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
	"github.com/iost-official/go-iost/p2p"
)

var (
	metricsGeneratedBlockCount = metrics.NewCounter("iost_pob_generated_block", nil)
	metricsVerifyBlockCount    = metrics.NewCounter("iost_pob_verify_block", nil)
	metricsConfirmedLength     = metrics.NewGauge("iost_pob_confirmed_length", nil)
	metricsTxSize              = metrics.NewGauge("iost_block_tx_size", nil)
	metricsMode                = metrics.NewGauge("iost_node_mode", nil)
)

var (
	errSingle    = errors.New("single block")
	errDuplicate = errors.New("duplicate block")
	// errTxHash     = errors.New("wrong txs hash")
	// errMerkleHash = errors.New("wrong tx receipt merkle hash")
)

var blockReqTimeout = 3 * time.Second

type verifyBlockMessage struct {
	blk     *block.Block
	gen     bool
	p2pType p2p.MessageType
}

//PoB is a struct that handles the consensus logic.
type PoB struct {
	account         *account.KeyPair
	baseVariable    global.BaseVariable
	blockChain      block.Chain
	blockCache      blockcache.BlockCache
	txPool          txpool.TxPool
	p2pService      p2p.Service
	verifyDB        db.MVCCDB
	produceDB       db.MVCCDB
	blockReqMap     *sync.Map
	exitSignal      chan struct{}
	chRecvBlock     chan p2p.IncomingMessage
	chRecvBlockHash chan p2p.IncomingMessage
	chQueryBlock    chan p2p.IncomingMessage
	chVerifyBlock   chan *verifyBlockMessage
}

// New init a new PoB.
func New(account *account.KeyPair, baseVariable global.BaseVariable, blockCache blockcache.BlockCache, txPool txpool.TxPool, p2pService p2p.Service) *PoB {
	p := PoB{
		account:         account,
		baseVariable:    baseVariable,
		blockChain:      baseVariable.BlockChain(),
		blockCache:      blockCache,
		txPool:          txPool,
		p2pService:      p2pService,
		verifyDB:        baseVariable.StateDB(),
		produceDB:       baseVariable.StateDB().Fork(),
		blockReqMap:     new(sync.Map),
		exitSignal:      make(chan struct{}),
		chRecvBlock:     p2pService.Register("consensus channel", p2p.NewBlock, p2p.SyncBlockResponse),
		chRecvBlockHash: p2pService.Register("consensus block head", p2p.NewBlockHash),
		chQueryBlock:    p2pService.Register("consensus query block", p2p.NewBlockRequest),
		chVerifyBlock:   make(chan *verifyBlockMessage, 1024),
	}
	staticProperty = newStaticProperty(p.account, blockCache.LinkedRoot().Active())

	// exec vote_producer
	act := tx.NewAction("iost.base", "Exec", fmt.Sprintf(`[]`))
	verifier.BlockBaseTx.Actions = append(verifier.BlockBaseTx.Actions, act)

	return &p
}

//Start make the PoB run.
func (p *PoB) Start() error {
	go p.messageLoop()
	go p.blockLoop()
	go p.verifyLoop()
	go p.scheduleLoop()
	return nil
}

//Stop make the PoB stop
func (p *PoB) Stop() {
	close(p.exitSignal)
}

func (p *PoB) messageLoop() {
	for {
		if p.baseVariable.Mode() != global.ModeInit {
			break
		}
		time.Sleep(time.Second)
	}
	for {
		select {
		case incomingMessage, ok := <-p.chRecvBlockHash:
			if !ok {
				ilog.Infof("chRecvBlockHash has closed")
				return
			}
			if p.baseVariable.Mode() == global.ModeNormal {
				var blkInfo msgpb.BlockInfo
				err := blkInfo.Unmarshal(incomingMessage.Data())
				if err != nil {
					continue
				}
				p.handleRecvBlockHash(&blkInfo, incomingMessage.From())
			}
		case incomingMessage, ok := <-p.chQueryBlock:
			if !ok {
				ilog.Infof("chQueryBlock has closed")
				return
			}
			if p.baseVariable.Mode() == global.ModeNormal {
				var rh msgpb.BlockInfo
				err := rh.Unmarshal(incomingMessage.Data())
				if err != nil {
					continue
				}
				p.handleBlockQuery(&rh, incomingMessage.From())
			}
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) handleRecvBlockHash(blkInfo *msgpb.BlockInfo, peerID p2p.PeerID) {
	_, ok := p.blockReqMap.Load(string(blkInfo.Hash))
	if ok {
		ilog.Info("block in block request map, block number: ", blkInfo.Number)
		return
	}
	_, err := p.blockCache.Find(blkInfo.Hash)
	if err == nil {
		ilog.Info("duplicate block, block number: ", blkInfo.Number)
		return
	}
	bytes, err := blkInfo.Marshal()
	if err != nil {
		ilog.Debugf("fail to Marshal requestblock, %v", err)
		return
	}
	p.blockReqMap.Store(string(blkInfo.Hash), time.AfterFunc(blockReqTimeout, func() {
		p.blockReqMap.Delete(string(blkInfo.Hash))
	}))
	p.p2pService.SendToPeer(peerID, bytes, p2p.NewBlockRequest, p2p.UrgentMessage, true)
}

func (p *PoB) handleBlockQuery(rh *msgpb.BlockInfo, peerID p2p.PeerID) {
	var blk *block.Block
	node, err := p.blockCache.Find(rh.Hash)
	if err == nil {
		blk = node.Block
	} else {
		blk, err = p.baseVariable.BlockChain().GetBlockByHash(rh.Hash)
		if err != nil {
			ilog.Errorf("handle block query failed to get block.")
			return
		}
	}
	b, err := blk.Encode()
	if err != nil {
		ilog.Errorf("Fail to encode block: %v, err=%v", rh.Number, err)
		return
	}
	p.p2pService.SendToPeer(peerID, b, p2p.NewBlock, p2p.UrgentMessage, true)
}

func (p *PoB) broadcastBlockHash(blk *block.Block) {
	blkInfo := &msgpb.BlockInfo{
		Number: blk.Head.Number,
		Hash:   blk.HeadHash(),
	}
	b, err := blkInfo.Marshal()
	if err != nil {
		ilog.Error("fail to encode block hash")
	} else {
		if p.baseVariable.Mode() == global.ModeNormal {
			p.p2pService.Broadcast(b, p2p.NewBlockHash, p2p.UrgentMessage, true)
		}
	}
}

func (p *PoB) doVerifyBlock(vbm *verifyBlockMessage) {
	if p.baseVariable.Mode() == global.ModeInit {
		return
	}
	ilog.Infof("verify block chan size:%v", len(p.chVerifyBlock))
	blk := vbm.blk
	if vbm.gen {
		ilog.Info("block from myself, block number: ", blk.Head.Number)
		err := p.handleRecvBlock(blk)
		if err != nil {
			ilog.Errorf("received block from myself, error, err:%v", err)
		}
		return
	}
	switch vbm.p2pType {
	case p2p.NewBlock:
		ilog.Info("received new block, block number: ", blk.Head.Number)
		timer, ok := p.blockReqMap.Load(string(blk.HeadHash()))
		if ok {
			t, ok := timer.(*time.Timer)
			if ok {
				t.Stop()
			}
		} else {
			p.blockReqMap.Store(string(blk.HeadHash()), nil)
		}
		err := p.handleRecvBlock(blk)
		p.broadcastBlockHash(blk) // can use go
		p.blockReqMap.Delete(string(blk.HeadHash()))
		if err != nil {
			ilog.Errorf("received new block error, err:%v", err)
			return
		}
	case p2p.SyncBlockResponse:
		ilog.Info("received sync block, block number: ", blk.Head.Number)
		err := p.handleRecvBlock(blk)
		if err != nil {
			ilog.Errorf("received sync block error, err:%v", err)
			return
		}
	}
	metricsVerifyBlockCount.Add(1, nil)
}

func (p *PoB) verifyLoop() {
	for {
		select {
		case vbm := <-p.chVerifyBlock:
			p.doVerifyBlock(vbm)
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) blockLoop() {
	ilog.Infof("start blockloop")
	for {
		select {
		case incomingMessage, ok := <-p.chRecvBlock:
			if !ok {
				ilog.Infof("chRecvBlock has closed")
				return
			}
			ilog.Infof("recv block chan size:%v", len(p.chRecvBlock))
			var blk block.Block
			err := blk.Decode(incomingMessage.Data())
			if err != nil {
				ilog.Error("fail to decode block")
				continue
			}
			ilog.Info("received block, block number: ", blk.Head.Number)
			p.chVerifyBlock <- &verifyBlockMessage{blk: &blk, gen: false, p2pType: incomingMessage.Type()}
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
			ilog.Info(p.baseVariable.Mode())
			metricsMode.Set(float64(p.baseVariable.Mode()), nil)
			if witnessOfSec(time.Now().Unix()) == p.account.ID {
				if p.baseVariable.Mode() == global.ModeNormal {
					p.txPool.Lock()
					blk, err := generateBlock(p.account, p.txPool, p.produceDB)
					p.txPool.Release()
					ilog.Infof("gen block:%v", blk.Head.Number)
					if err != nil {
						ilog.Error(err.Error())
						continue
					}
					ilog.Debugf("block tx num: %v", len(blk.Txs))
					p.chVerifyBlock <- &verifyBlockMessage{blk: blk, gen: true}
					blkByte, err := blk.Encode()
					if err != nil {
						ilog.Error(err.Error())
						continue
					}
					go p.p2pService.Broadcast(blkByte, p2p.NewBlock, p2p.UrgentMessage, true)
				}
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
		return errDuplicate
	}
	err = verifyBasics(blk.Head, blk.Sign)
	if err != nil {
		return err
	}
	parent, err := p.blockCache.Find(blk.Head.ParentHash)
	p.blockCache.Add(blk)
	if err == nil && parent.Type == blockcache.Linked {
		return p.addExistingBlock(blk, parent.Block)
	}
	return errSingle
}

func (p *PoB) addExistingBlock(blk *block.Block, parentBlock *block.Block) error {
	node, _ := p.blockCache.Find(blk.HeadHash())
	ok := p.verifyDB.Checkout(string(blk.HeadHash()))
	if !ok {
		p.verifyDB.Checkout(string(blk.Head.ParentHash))
		err := verifyBlock(blk, parentBlock, p.blockCache.LinkedRoot().Block, p.txPool, p.verifyDB, p.blockChain)
		if err != nil {
			ilog.Errorf("verify block failed. err=%v", err)
			p.blockCache.Del(node)
			return err
		}
		p.verifyDB.Tag(string(blk.HeadHash()))
	}
	h := p.blockCache.Head()
	if node.Head.Number > h.Head.Number {
		p.txPool.AddLinkedNode(node, node)
	} else {
		p.txPool.AddLinkedNode(node, h)
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
	staticProperty.updateWitness(p.blockCache.LinkedRoot().Active())
	if staticProperty.isWitness(p.account.ID) {
		p.p2pService.ConnectBPs(p.blockCache.LinkedRoot().NetID())
	}
}
