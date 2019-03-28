package pob

import (
	"sync"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/chainbase"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/consensus/synchro"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/db"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
	"github.com/iost-official/go-iost/p2p"
)

var (
	generateBlockCount         = metrics.NewCounter("iost_pob_generated_block", nil)
	verifyBlockCount           = metrics.NewCounter("iost_pob_verify_block", nil)
	generateBlockTimeGauge     = metrics.NewGauge("iost_pob_generate_block_time", nil)
	verifyBlockTimeGauge       = metrics.NewGauge("iost_pob_verify_block_time", nil)
	receiveBlockDelayTimeGauge = metrics.NewGauge("iost_pob_receive_block_delay_time", nil)
	metricsConfirmedLength     = metrics.NewGauge("iost_pob_confirmed_length", nil)
	metricsMode                = metrics.NewGauge("iost_node_mode", nil)
)

var (
	maxBlockNumber    = int64(10000)
	subSlotTime       = 500 * time.Millisecond
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
	verifyDB   db.MVCCDB
	produceDB  db.MVCCDB
	sync       *synchro.Sync
	cBase      *chainbase.ChainBase

	exitSignal       chan struct{}
	quitGenerateMode chan struct{}
	wg               *sync.WaitGroup
	mu               *sync.RWMutex
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

	// TODO: Organize the owner and lifecycle of all metrics.
	metricsMode.Set(float64(2), nil)

	p := PoB{
		account:    account,
		conf:       conf,
		blockChain: cBase.BlockChain(),
		blockCache: cBase.BlockCache(),
		txPool:     txPool,
		p2pService: p2pService,
		verifyDB:   cBase.StateDB(),
		produceDB:  cBase.StateDB().Fork(),
		sync:       nil,
		cBase:      cBase,

		exitSignal:       make(chan struct{}),
		quitGenerateMode: make(chan struct{}),
		wg:               new(sync.WaitGroup),
		mu:               new(sync.RWMutex),
	}

	p.recoverBlockcache()
	close(p.quitGenerateMode)

	return &p
}

func (p *PoB) recoverBlockcache() error {
	err := p.blockCache.Recover(p.cBase)
	if err != nil {
		ilog.Error("Failed to recover blockCache, err: ", err)
		ilog.Info("Don't Recover, Move old file to BlockCacheWALCorrupted")
		err = p.blockCache.NewWAL(p.conf)
		if err != nil {
			ilog.Error(" Failed to NewWAL, err: ", err)
		}
	}
	return err
}

// Start make the PoB run.
func (p *PoB) Start() error {
	p.sync = synchro.New(p.p2pService, p.blockCache, p.blockChain)

	p.wg.Add(2)
	go p.verifyLoop()
	go p.scheduleLoop()
	return nil
}

// Stop make the PoB stop.
func (p *PoB) Stop() {
	close(p.exitSignal)
	p.wg.Wait()

	p.sync.Close()
}

// Mode return the mode of pob.
func (p *PoB) Mode() string {
	if p.sync == nil {
		return "ModeInit"
	} else if p.sync.IsCatchingUp() {
		return "ModeSync"
	} else {
		return "ModeNormal"
	}
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
		if err != chainbase.ErrSingle && err != chainbase.ErrDuplicate {
			ilog.Warnf("Verify block failed: %v", err)
		}
		return
	}

	metricsConfirmedLength.Set(float64(p.blockCache.LinkedRoot().Head.Number), nil)

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
			select {
			case <-p.quitGenerateMode:
			}
			p.doVerifyBlock(blk)
		case <-p.exitSignal:
			p.wg.Done()
			return
		}
	}
}

func (p *PoB) scheduleLoop() {
	defer p.wg.Done()
	nextSchedule := common.TimeUntilNextSchedule(time.Now().UnixNano())
	ilog.Debugf("nextSchedule: %.2f", time.Duration(nextSchedule).Seconds())
	pubkey := p.account.ReadablePubkey()

	var slotFlag int64
	for {
		select {
		case <-time.After(time.Duration(nextSchedule)):
			time.Sleep(time.Millisecond)
			if p.sync.IsCatchingUp() {
				metricsMode.Set(float64(1), nil)
			} else {
				metricsMode.Set(float64(0), nil)
			}
			t := time.Now()
			pTx, head := p.txPool.PendingTx()
			witnessList := head.Active()
			if slotFlag != common.SlotOfNanoSec(t.UnixNano()) && !p.sync.IsCatchingUp() && common.WitnessOfNanoSec(t.UnixNano(), witnessList) == pubkey {
				p.quitGenerateMode = make(chan struct{})
				slotFlag = common.SlotOfNanoSec(t.UnixNano())
				generateBlockTicker := time.NewTicker(subSlotTime)
				for num := 0; num < common.BlockNumPerWitness; num++ {
					p.gen(num, pTx, head)
					if num == common.BlockNumPerWitness-1 {
						break
					}
					select {
					case <-generateBlockTicker.C:
					}
					pTx, head = p.txPool.PendingTx()
					witnessList = head.Active()
					if common.WitnessOfNanoSec(time.Now().UnixNano(), witnessList) != pubkey {
						break
					}
				}
				close(p.quitGenerateMode)
				generateBlockTicker.Stop()
			}
			nextSchedule = common.TimeUntilNextSchedule(time.Now().UnixNano())
			ilog.Debugf("nextSchedule: %.2f", time.Duration(nextSchedule).Seconds())
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) gen(num int, pTx *txpool.SortedTxMap, head *blockcache.BlockCacheNode) {
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
	p.txPool.Lock()
	blk, err := generateBlock(p.account, p.txPool, p.produceDB, limitTime, pTx, head)
	p.txPool.Release()
	if err != nil {
		ilog.Error(err)
		return
	}

	blkByte, err := blk.Encode()
	if err != nil {
		ilog.Error(err)
		return
	}
	p.p2pService.Broadcast(blkByte, p2p.NewBlock, p2p.UrgentMessage)

	p.mu.Lock()
	err = p.cBase.Add(blk, false, true)
	p.mu.Unlock()
	if err != nil {
		ilog.Errorf("[pob] handle block from myself, err:%v", err)
		return
	}
	metricsConfirmedLength.Set(float64(p.blockCache.LinkedRoot().Head.Number), nil)
}
