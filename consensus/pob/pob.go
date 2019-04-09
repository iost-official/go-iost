package pob

import (
	"fmt"
	"sync"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/chainbase"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/synchro"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/verifier"
)

var (
	generateBlockCount         = metrics.NewCounter("iost_pob_generated_block", nil)
	verifyBlockCount           = metrics.NewCounter("iost_pob_verify_block", nil)
	generateBlockTimeGauge     = metrics.NewGauge("iost_pob_generate_block_time", nil)
	verifyBlockTimeGauge       = metrics.NewGauge("iost_pob_verify_block_time", nil)
	receiveBlockDelayTimeGauge = metrics.NewGauge("iost_pob_receive_block_delay_time", nil)
	metricsConfirmedLength     = metrics.NewGauge("iost_pob_confirmed_length", nil)
)

var (
	maxBlockNumber    = int64(10000)
	last2GenBlockTime = 50 * time.Millisecond
)

//PoB is a struct that handles the consensus logic.
type PoB struct {
	account    *account.KeyPair
	conf       *common.Config
	blockChain block.Chain
	blockCache blockcache.BlockCache
	txPool     txpool.TxPool
	p2pService p2p.Service
	produceDB  db.MVCCDB
	sync       *synchro.Sync
	cBase      *chainbase.ChainBase

	exitSignal chan struct{}
	wg         *sync.WaitGroup
	mu         *sync.RWMutex
}

// New init a new PoB.
func New(conf *common.Config, cBase *chainbase.ChainBase, txPool txpool.TxPool, p2pService p2p.Service) *PoB {
	// TODO: Move the code to account struct.
	accSecKey := conf.ACC.SecKey
	accAlgo := conf.ACC.Algorithm
	account, err := account.NewKeyPair(common.Base58Decode(accSecKey), crypto.NewAlgorithm(accAlgo))
	if err != nil {
		ilog.Fatalf("NewKeyPair failed, stop the program! err:%v", err)
	}

	p := PoB{
		account:    account,
		conf:       conf,
		blockChain: cBase.BlockChain(),
		blockCache: cBase.BlockCache(),
		txPool:     txPool,
		p2pService: p2pService,
		produceDB:  cBase.StateDB().Fork(),
		sync:       nil,
		cBase:      cBase,

		exitSignal: make(chan struct{}),
		wg:         new(sync.WaitGroup),
		mu:         new(sync.RWMutex),
	}

	return &p
}

// Start make the PoB run.
func (p *PoB) Start() error {
	p.sync = synchro.New(p.p2pService, p.blockCache, p.blockChain)

	p.wg.Add(3)
	go p.verifyLoop()
	go p.generateLoop()
	go p.metricsLoop()
	return nil
}

// Stop make the PoB stop.
func (p *PoB) Stop() {
	close(p.exitSignal)
	p.wg.Wait()

	p.sync.Close()
}

func (p *PoB) doVerifyBlock(blk *block.Block) {
	now := time.Now().UnixNano()
	receiveBlockDelayTimeGauge.Set(float64(now-blk.Head.Time), nil)
	defer func() {
		verifyBlockTimeGauge.Set(float64(time.Now().UnixNano()-now), nil)
		verifyBlockCount.Add(1, nil)
	}()

	head := p.blockCache.Head().Head.Number
	if blk.Head.Number > head+maxBlockNumber {
		ilog.Debugf("Block number %v is %v higher than head number %v", blk.Head.Number, maxBlockNumber, head)
		return
	}

	p.mu.Lock()
	err := p.cBase.Add(blk, false, false)
	p.mu.Unlock()
	if err != nil {
		return
	}

	if common.IsWitness(p.account.ReadablePubkey(), p.blockCache.Head().Active()) {
		p.p2pService.ConnectBPs(p.blockCache.Head().NetID())
	} else {
		p.p2pService.ConnectBPs(nil)
	}

	if !p.sync.IsCatchingUp() {
		p.sync.BroadcastBlockInfo(blk)
	}
}

func (p *PoB) verifyLoop() {
	for {
		select {
		case blk := <-p.sync.IncomingBlock():
			p.doVerifyBlock(blk)
		case <-p.exitSignal:
			p.wg.Done()
			return
		}
	}
}

func (p *PoB) doGenerateBlock(slot int64) {
	// When the iserver is catching up, the generate block is not performed.
	if p.sync.IsCatchingUp() {
		return
	}

	p.mu.Lock()
	for num := 0; num < common.BlockNumPerWitness; num++ {
		<-time.After(time.Until(common.TimeOfBlock(slot, int64(num))))
		witnessList := p.blockCache.Head().Active()
		if common.WitnessOfNanoSec(time.Now().UnixNano(), witnessList) != p.account.ReadablePubkey() {
			break
		}
		p.gen(num)
	}
	p.mu.Unlock()
}

func (p *PoB) generateLoop() {
	for {
		slot := common.NextSlot()
		select {
		case <-time.After(time.Until(common.TimeOfBlock(slot, 0))):
			p.doGenerateBlock(slot)
		case <-p.exitSignal:
			p.wg.Done()
			return
		}
	}
}

func (p *PoB) gen(num int) {
	now := time.Now().UnixNano()
	defer func() {
		// TODO: Confirm the most appropriate metrics definition.
		generateBlockTimeGauge.Set(float64(time.Now().UnixNano()-now), nil)
		generateBlockCount.Add(1, nil)
	}()

	limitTime := common.MaxBlockTimeLimit
	if num >= common.BlockNumPerWitness-2 {
		limitTime = last2GenBlockTime
	}
	blk, err := p.generateBlock(limitTime)
	if err != nil {
		ilog.Error(err)
		return
	}

	p.sync.BroadcastBlock(blk)

	err = p.cBase.Add(blk, false, true)
	if err != nil {
		ilog.Errorf("[pob] handle block from myself, err:%v", err)
		return
	}
}

func (p *PoB) generateBlock(limitTime time.Duration) (*block.Block, error) {
	st := time.Now()
	pTx, head := p.txPool.PendingTx()
	witnessList := head.Active()
	if common.WitnessOfNanoSec(st.UnixNano(), witnessList) != p.account.ReadablePubkey() {
		return nil, fmt.Errorf("Now time %v exceeding the slot of witness %v", st.UnixNano(), p.account.ReadablePubkey())
	}
	blk := &block.Block{
		Head: &block.BlockHead{
			Version:    0,
			ParentHash: head.HeadHash(),
			Info:       make([]byte, 0),
			Number:     head.Head.Number + 1,
			Witness:    p.account.ReadablePubkey(),
			Time:       time.Now().UnixNano(),
		},
		Txs:      []*tx.Tx{},
		Receipts: []*tx.TxReceipt{},
	}
	p.produceDB.Checkout(string(head.HeadHash()))

	// call vote
	v := verifier.Verifier{}
	// TODO: stateDb and block head is consisdent, pTx may be inconsisdent.
	dropList, _, err := v.Gen(blk, head.Block, &head.WitnessList, p.produceDB, pTx, &verifier.Config{
		Mode:        0,
		Timeout:     limitTime - time.Now().Sub(st),
		TxTimeLimit: common.MaxTxTimeLimit,
	})
	if err != nil {
		go p.delTxList(dropList)
		ilog.Errorf("Gen is err: %v", err)
		return nil, err
	}
	blk.Head.TxMerkleHash = blk.CalculateTxMerkleHash()
	blk.Head.TxReceiptMerkleHash = blk.CalculateTxReceiptMerkleHash()
	err = blk.CalculateHeadHash()
	if err != nil {
		return nil, err
	}
	blk.Sign = p.account.Sign(blk.HeadHash())
	p.produceDB.Commit(string(blk.HeadHash()))
	return blk, nil
}

func (p *PoB) delTxList(delList []*tx.Tx) {
	for _, t := range delList {
		p.txPool.DelTx(t.Hash())
	}
}

func (p *PoB) metricsLoop() {
	for {
		select {
		case <-time.After(2 * time.Second):
			if p.sync.IsCatchingUp() {
				common.SetMode(common.ModeSync)
			} else {
				common.SetMode(common.ModeNormal)
			}
			metricsConfirmedLength.Set(float64(p.blockCache.LinkedRoot().Head.Number), nil)
		case <-p.exitSignal:
			p.wg.Done()
			return
		}
	}
}
