package pob

import (
	"errors"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/snapshot"
	msgpb "github.com/iost-official/go-iost/consensus/synchronizer/pb"
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
	metricsGeneratedBlockCount   = metrics.NewCounter("iost_pob_generated_block", nil)
	metricsVerifyBlockCount      = metrics.NewCounter("iost_pob_verify_block", nil)
	metricsConfirmedLength       = metrics.NewGauge("iost_pob_confirmed_length", nil)
	metricsTxSize                = metrics.NewGauge("iost_block_tx_size", nil)
	metricsMode                  = metrics.NewGauge("iost_node_mode", nil)
	metricsTimeCost              = metrics.NewGauge("iost_time_cost", nil)
	metricsTransferCost          = metrics.NewGauge("iost_transfer_cost", nil)
	metricsGenerateBlockTimeCost = metrics.NewGauge("iost_generate_block_time_cost", nil)
	metricsDelayedBlock          = metrics.NewCounter("iost_delayed_block", nil)
)

var (
	errSingle    = errors.New("single block")
	errDuplicate = errors.New("duplicate block")
)

var (
	blockReqTimeout   = 3 * time.Second
	continuousNum     int
	subSlotTime       = 300 * time.Millisecond
	genBlockTime      = 250 * time.Millisecond
	last2GenBlockTime = 30 * time.Millisecond
	tWitness          = ""
	tContinuousNum    = 0
)

type verifyBlockMessage struct {
	blk     *block.Block
	p2pType p2p.MessageType
}

//PoB is a struct that handles the consensus logic.
type PoB struct {
	account          *account.KeyPair
	baseVariable     global.BaseVariable
	blockChain       block.Chain
	blockCache       blockcache.BlockCache
	txPool           txpool.TxPool
	p2pService       p2p.Service
	verifyDB         db.MVCCDB
	produceDB        db.MVCCDB
	blockReqMap      *sync.Map
	exitSignal       chan struct{}
	quitGenerateMode chan struct{}
	chRecvBlock      chan p2p.IncomingMessage
	chRecvBlockHash  chan p2p.IncomingMessage
	chQueryBlock     chan p2p.IncomingMessage
	chVerifyBlock    chan *verifyBlockMessage
	wg               *sync.WaitGroup
	mu               *sync.RWMutex
}

// New init a new PoB.
func New(account *account.KeyPair, baseVariable global.BaseVariable, blockCache blockcache.BlockCache, txPool txpool.TxPool, p2pService p2p.Service) *PoB {
	p := PoB{
		account:          account,
		baseVariable:     baseVariable,
		blockChain:       baseVariable.BlockChain(),
		blockCache:       blockCache,
		txPool:           txPool,
		p2pService:       p2pService,
		verifyDB:         baseVariable.StateDB(),
		produceDB:        baseVariable.StateDB().Fork(),
		blockReqMap:      new(sync.Map),
		exitSignal:       make(chan struct{}),
		quitGenerateMode: make(chan struct{}),
		chRecvBlock:      p2pService.Register("consensus channel", p2p.NewBlock, p2p.SyncBlockResponse),
		chRecvBlockHash:  p2pService.Register("consensus block head", p2p.NewBlockHash),
		chQueryBlock:     p2pService.Register("consensus query block", p2p.NewBlockRequest),
		chVerifyBlock:    make(chan *verifyBlockMessage, 1024),
		wg:               new(sync.WaitGroup),
		mu:               new(sync.RWMutex),
	}
	continuousNum = baseVariable.Continuous()
	staticProperty = newStaticProperty(p.account, blockCache.LinkedRoot().PendingNum())
	p.recoverBlockcache()
	close(p.quitGenerateMode)
	return &p
}

func (p *PoB) recoverBlockcache() error {
	err := p.blockCache.Recover(p)
	if err != nil {
		ilog.Error("Failed to recover blockCache, err: ", err)
		ilog.Info("Don't Recover, Move old file to BlockCacheWALCorrupted")
		err = p.blockCache.NewWAL(p.baseVariable.Config())
		if err != nil {
			ilog.Error(" Failed to NewWAL, err: ", err)
		}
	}
	return err
}

//Start make the PoB run.
func (p *PoB) Start() error {

	p.wg.Add(4)
	go p.messageLoop()
	go p.blockLoop()
	go p.verifyLoop()
	go p.scheduleLoop()
	return nil
}

//Stop make the PoB stop
func (p *PoB) Stop() {
	close(p.exitSignal)
	p.wg.Wait()
}

func (p *PoB) messageLoop() {
	defer p.wg.Done()
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
				err := proto.Unmarshal(incomingMessage.Data(), &blkInfo)
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
				err := proto.Unmarshal(incomingMessage.Data(), &rh)
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
		//ilog.Debug("block in block request map, block number: ", blkInfo.Number)
		return
	}
	_, err := p.blockCache.Find(blkInfo.Hash)
	if err == nil {
		ilog.Debug("duplicate block, block number: ", blkInfo.Number)
		return
	}
	bytes, err := proto.Marshal(blkInfo)
	if err != nil {
		ilog.Debugf("fail to Marshal requestblock, %v", err)
		return
	}
	p.blockReqMap.Store(string(blkInfo.Hash), time.AfterFunc(blockReqTimeout, func() {
		p.blockReqMap.Delete(string(blkInfo.Hash))
	}))
	p.p2pService.SendToPeer(peerID, bytes, p2p.NewBlockRequest, p2p.UrgentMessage)
}

func (p *PoB) handleBlockQuery(rh *msgpb.BlockInfo, peerID p2p.PeerID) {
	var blk *block.Block
	blk, err := p.blockCache.GetBlockByHash(rh.Hash)
	if err != nil {
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
	p.p2pService.SendToPeer(peerID, b, p2p.NewBlock, p2p.UrgentMessage)
}

func (p *PoB) broadcastBlockHash(blk *block.Block) {
	if p.baseVariable.Mode() != global.ModeNormal {
		return
	}

	blkInfo := &msgpb.BlockInfo{
		Number: blk.Head.Number,
		Hash:   blk.HeadHash(),
	}
	b, err := proto.Marshal(blkInfo)
	if err != nil {
		ilog.Errorf("fail to encode block hash, err=%v, blockHash=%+v", err, *blkInfo)
	} else {
		p.p2pService.Broadcast(b, p2p.NewBlockHash, p2p.UrgentMessage)
	}
}

func calculateTime(blk *block.Block) float64 {
	return float64((time.Now().UnixNano() - blk.Head.Time) / 1e6)
}

func (p *PoB) doVerifyBlock(vbm *verifyBlockMessage) {
	if p.baseVariable.Mode() == global.ModeInit {
		return
	}
	ilog.Debugf("verify block chan size:%v", len(p.chVerifyBlock))
	blk := vbm.blk
	switch vbm.p2pType {
	case p2p.NewBlock:
		t1 := calculateTime(blk)
		metricsTransferCost.Set(t1, nil)
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
		t2 := calculateTime(blk)
		metricsTimeCost.Set(t2, nil)
		if err == errSingle || err == nil {
			go p.broadcastBlockHash(blk)
		}
		p.blockReqMap.Delete(string(blk.HeadHash()))
		if err != nil && err != errSingle && err != errDuplicate {
			ilog.Warnf("received new block error, err:%v", err)
			return
		}
	case p2p.SyncBlockResponse:
		err := p.handleRecvBlock(blk)
		if err != nil && err != errSingle && err != errDuplicate {
			ilog.Warnf("received sync block error, err:%v", err)
			return
		}
	}
	metricsVerifyBlockCount.Add(1, nil)
}

func (p *PoB) verifyLoop() {
	defer p.wg.Done()
	for {
		select {
		case vbm := <-p.chVerifyBlock:
			select {
			case <-p.quitGenerateMode:
			}
			p.doVerifyBlock(vbm)
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) blockLoop() {
	ilog.Infof("start blockloop")
	defer p.wg.Done()
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
				ilog.Error("fail to decode block")
				continue
			}
			p.chVerifyBlock <- &verifyBlockMessage{blk: &blk, p2pType: incomingMessage.Type()}
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) scheduleLoop() {
	defer p.wg.Done()

	tGenBlock := time.NewTicker(20 * time.Millisecond)
	tMetricsMode := time.NewTicker(3 * time.Second)
	defer tGenBlock.Stop()
	defer tMetricsMode.Stop()

	var slotFlag int64
	for {
		select {
		case <-tGenBlock.C:
			// Don't delete,avoid time error
			time.Sleep(1 * time.Millisecond)

			t := time.Now()
			_, head := p.txPool.PendingTx()
			witnessList := head.Active()
			pubkey := p.account.ReadablePubkey()
			if witnessOfNanoSec(t.UnixNano(), witnessList) == pubkey && slotFlag != slotOfSec(t.Unix()) && p.baseVariable.Mode() == global.ModeNormal {
				slotFlag = slotOfSec(t.Unix())
				p.quitGenerateMode = make(chan struct{})
				generateBlockTicker := time.NewTicker(subSlotTime)
				generateTxsNum = 0
				for num := 0; num < continuousNum; num++ {
					p.gen(num)
					if num == continuousNum-1 {
						break
					}
					select {
					case <-generateBlockTicker.C:
					}
					if witnessOfNanoSec(t.UnixNano(), witnessList) != pubkey {
						break
					}
				}
				close(p.quitGenerateMode)
				metricsTxSize.Set(float64(generateTxsNum), nil)
				generateBlockTicker.Stop()
			}
		case <-tMetricsMode.C:
			metricsMode.Set(float64(p.baseVariable.Mode()), nil)
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) gen(num int) {
	limitTime := genBlockTime
	if num >= continuousNum-2 {
		limitTime = last2GenBlockTime
	}
	p.txPool.Lock()
	blk, err := generateBlock(p.account, p.txPool, p.produceDB, limitTime)
	p.txPool.Release()
	if err != nil {
		ilog.Error(err)
		return
	}
	p.printStatistics(num, blk)
	blkByte, err := blk.Encode()
	if err != nil {
		ilog.Error(err)
		return
	}
	p.p2pService.Broadcast(blkByte, p2p.NewBlock, p2p.UrgentMessage)
	metricsGenerateBlockTimeCost.Set(calculateTime(blk), nil)
	err = p.handleRecvBlock(blk)
	if err != nil {
		ilog.Errorf("[pob] handle block from myself, err:%v", err)
		return
	}
}

func (p *PoB) printStatistics(num int, blk *block.Block) {
	ptx, _ := p.txPool.PendingTx()
	ilog.Infof("Gen block - @%v id:%v..., t:%v, num:%v, confirmed:%v, txs:%v, pendingtxs:%v, et:%vms",
		num,
		p.account.ReadablePubkey()[:10],
		blk.Head.Time,
		blk.Head.Number,
		p.blockCache.LinkedRoot().Head.Number,
		len(blk.Txs),
		ptx.Size(),
		calculateTime(blk),
	)
}

// RecoverBlock recover block from block cache wal
func (p *PoB) RecoverBlock(blk *block.Block, witnessList blockcache.WitnessList) error {
	_, err := p.blockCache.Find(blk.HeadHash())
	if err == nil {
		return errDuplicate
	}
	err = verifyBasics(blk, blk.Sign)
	if err != nil {
		return err
	}
	parent, err := p.blockCache.Find(blk.Head.ParentHash)
	p.blockCache.AddWithWit(blk, witnessList)
	if err == nil && parent.Type == blockcache.Linked {
		return p.addExistingBlock(blk, parent.Block, true)
	}
	return errSingle
}

func (p *PoB) handleRecvBlock(blk *block.Block) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, err := p.blockCache.Find(blk.HeadHash())
	if err == nil {
		return errDuplicate
	}
	err = verifyBasics(blk, blk.Sign)
	if err != nil {
		return err
	}
	parent, err := p.blockCache.Find(blk.Head.ParentHash)
	p.blockCache.Add(blk)
	if err == nil && parent.Type == blockcache.Linked {
		return p.addExistingBlock(blk, parent.Block, false)
	}
	return errSingle
}

func (p *PoB) addExistingBlock(blk *block.Block, parentBlock *block.Block, replay bool) error {
	node, _ := p.blockCache.Find(blk.HeadHash())
	ok := p.verifyDB.Checkout(string(blk.HeadHash()))
	if !ok {
		p.verifyDB.Checkout(string(blk.Head.ParentHash))
		p.txPool.Lock()
		err := verifyBlock(blk, parentBlock, p.blockCache.LinkedRoot().Block, &node.GetParent().WitnessList, p.txPool, p.verifyDB, p.blockChain, replay)
		p.txPool.Release()
		if err != nil {
			ilog.Errorf("verify block failed, blockNum:%v, blockHash:%v. err=%v", blk.Head.Number, common.Base58Encode(blk.HeadHash()), err)
			p.blockCache.Del(node)
			return err
		}
		err = snapshot.Save(p.verifyDB, blk)
		if err != nil {
			return err
		}
		p.verifyDB.Commit(string(blk.HeadHash()))
	}
	p.txPool.AddLinkedNode(node)
	p.blockCache.Link(node)
	p.updateInfo(node)
	if node.Head.Witness != p.account.ReadablePubkey() {
		if tWitness != node.Head.Witness {
			tWitness = node.Head.Witness
			tContinuousNum = 0
		}
		ilog.Infof("Rec block - @%v id:%v..., num:%v, t:%v, txs:%v, confirmed:%v, et:%vms",
			tContinuousNum, node.Head.Witness[:10], node.Head.Number, node.Head.Time, len(node.Txs), p.blockCache.LinkedRoot().Head.Number, calculateTime(node.Block))
		tContinuousNum++
	}
	if witnessOfNanoSec(time.Now().UnixNano(), node.GetParent().Active()) != node.Head.Witness {
		ilog.Debugf("hasn't process the block in the slot belonging to the witness")
		metricsDelayedBlock.Add(1, nil)
	}
	for child := range node.Children {
		p.addExistingBlock(child.Block, node.Block, replay)
	}
	return nil
}

func (p *PoB) updateInfo(node *blockcache.BlockCacheNode) {
	updateLib(node, p.blockCache)
	if staticProperty.isWitness(p.account.ReadablePubkey(), p.blockCache.LinkedRoot().Pending()) {
		p.p2pService.ConnectBPs(p.blockCache.LinkedRoot().NetID())
	}
}
