package pob

import (
	"encoding/binary"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/consensus/common"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/network"

	"errors"
	"fmt"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/message"

	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/core/txpool"
	"github.com/iost-official/Go-IOS-Protocol/log"
	"github.com/iost-official/Go-IOS-Protocol/verifier"
	"github.com/iost-official/Go-IOS-Protocol/vm"
	"github.com/iost-official/Go-IOS-Protocol/vm/lua"
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
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
			Name: "tx_pool_size",
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

var TxPerBlk int

type PoB struct {
	account      account.Account
	blockCache   blockcache.BlockCache
	router       network.Router
	synchronizer consensus_common.Synchronizer
	globalStaticProperty
	globalDynamicProperty

	exitSignal chan struct{}
	chBlock    chan message.Message

	log *log.Logger
}

func NewPoB(acc account.Account, bc block.Chain, pool state.Pool, witnessList []string) (*PoB, error) {
	TxPerBlk = 800
	p := PoB{
		account: acc,
	}

	p.blockCache = blockcache.NewBlockCache(bc, pool, len(witnessList)*2/3)
	if bc.GetBlockByNumber(0) == nil {
		t := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		err := p.genesis(consensus_common.GetTimestamp(t.Unix()).Slot)
		if err != nil {
			return nil, fmt.Errorf("failed to genesis is nil")
		}
	}

	var err error
	p.router = network.Route
	if p.router == nil {
		return nil, fmt.Errorf("failed to network.Route is nil")
	}

	p.synchronizer, err = consensus_common.NewSynchronizer(p.blockCache, p.router, len(witnessList)*2/3)

	p.chBlock, err = p.router.FilteredChan(network.Filter{
		AcceptType: []network.ReqType{network.ReqNewBlock, network.ReqSyncBlock}})
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

	p.update(&bc.Top().Head)
	return &p, nil
}

func (p *PoB) initGlobalProperty(acc account.Account, witnessList []string) {
	p.globalStaticProperty = newGlobalStaticProperty(acc, witnessList)
	p.globalDynamicProperty = newGlobalDynamicProperty()
}

func (p *PoB) Run() {
	p.synchronizer.StartListen()
	go p.blockLoop()
	go p.scheduleLoop()
}

func (p *PoB) Stop() {
	close(p.chBlock)
	close(p.exitSignal)
}

func (p *PoB) BlockCache() blockcache.BlockCache {
	return p.blockCache
}

func (p *PoB) BlockChain() block.Chain {
	return p.blockCache.BlockChain()
}

func (p *PoB) CachedBlockChain() block.Chain {
	return p.blockCache.LongestChain()
}

func (p *PoB) StatePool() state.Pool {
	return p.blockCache.BasePool()
}

func (p *PoB) CachedStatePool() state.Pool {
	return p.blockCache.LongestPool()
}

func (p *PoB) genesis(initTime int64) error {

	main := lua.NewMethod(vm.Public, "", 0, 0)

	var code string
	for k, v := range account.GenesisAccount {
		code += fmt.Sprintf("@PutHM iost %v f%v\n", k, v)
	}

	lc := lua.NewContract(vm.ContractInfo{Prefix: "", GasLimit: 0, Price: 0, Publisher: ""}, code, main)

	tmp_tx := tx.Tx{
		Time:     0,
		Nonce:    0,
		Contract: &lc,
	}

	genesis := &block.Block{
		Head: block.BlockHead{
			Version: 0,
			Number:  0,
			Time:    initTime,
		},
		Content: make([]tx.Tx, 0),
	}
	genesis.Content = append(genesis.Content, tmp_tx)
	stp, err := verifier.ParseGenesis(tmp_tx.Contract, p.StatePool())
	err = p.blockCache.SetBasePool(stp)
	err = p.blockCache.AddGenesis(genesis)
	return err
}

func (p *PoB) blockLoop() {
	for {
		select {
		case req, _ := <-p.chBlock:
			var blk block.Block
			blk.Decode(req.Body)
			localLength := p.blockCache.ConfirmedLength()
			if blk.Head.Number > int64(localLength)+consensus_common.MaxAcceptableLength {
				if req.ReqType == int32(network.ReqNewBlock){
					go p.synchronizer.SyncBlocks(localLength, localLength+uint64(consensus_common.MaxAcceptableLength))
				}
				continue
			}
			err := p.blockCache.Add(&blk, p.blockVerify)
			if err == nil {
				p.log.I("Link it onto cached chain")
				p.blockCache.SendOnBlock(&blk)
				receivedBlockCount.Inc()
			} else {
				p.log.I("Error: %v", err)
			}
			if err != blockcache.ErrBlock && err != blockcache.ErrTooOld {
				go p.synchronizer.BlockConfirmed(blk.Head.Number)
				if err == nil {
					p.globalDynamicProperty.update(&blk.Head)
				} else if err == blockcache.ErrNotFound && req.ReqType == int32(network.ReqNewBlock) {
					// New block is a single block
					need, start, end := p.synchronizer.NeedSync(uint64(blk.Head.Number))
					if need {
						go p.synchronizer.SyncBlocks(start, end)
					}
				}
			}
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) scheduleLoop() {
	var nextSchedule int64
	nextSchedule = 0
	for {
		select {
		case <-time.After(time.Second * time.Duration(nextSchedule)):
			currentTimestamp := consensus_common.GetCurrentTimestamp()
			wid := witnessOfTime(&p.globalStaticProperty, &p.globalDynamicProperty, currentTimestamp)
			if wid == p.account.ID {
				bc := p.blockCache.LongestChain()
				pool := p.blockCache.LongestPool()
				blk := p.genBlock(p.account, bc, pool)
				p.globalDynamicProperty.update(&blk.Head)
				bb := blk.Encode()
				msg := message.Message{ReqType: int32(network.ReqNewBlock), Body: bb}
				go p.router.Broadcast(msg)
				p.chBlock <- msg
				p.log.I("Broadcasted block, current timestamp: %v number: %v", currentTimestamp, blk.Head.Number)
			}
			nextSchedule = timeUntilNextSchedule(&p.globalStaticProperty, &p.globalDynamicProperty, time.Now().Unix())//当前结束到下一次造块的时间
			}
	}
}

func (p *PoB) genBlock(acc account.Account, bc block.Chain, pool state.Pool) *block.Block {
	lastBlk := bc.Top()
	blk := block.Block{Content: []tx.Tx{}, Head: block.BlockHead{
		Version:    0,
		ParentHash: lastBlk.HeadHash(),
		Number:     lastBlk.Head.Number + 1,
		Witness:    acc.ID,
		Time:       consensus_common.GetCurrentTimestamp().Slot,
	}}
	spool1 := pool.Copy()

	vc := vm.NewContext(vm.BaseContext())
	vc.Timestamp = blk.Head.Time
	vc.ParentHash = blk.Head.ParentHash
	vc.BlockHeight = blk.Head.Number
	vc.Witness = acc.ID

	txCnt := TxPerBlk + rand.Intn(500)
	var txList []*tx.Tx
	if txpool.TxPoolS != nil {
		txList = txpool.TxPoolS.PendingTransactions(txCnt)
		txPoolSize.Set(float64(txpool.TxPoolS.TransactionNum()))
	}

	if len(txList) != 0 {
		for _, t := range txList {
			blockcache.StdCacheVerifier(t, spool1, vc)
			blk.Content = append(blk.Content, *t)
		}
	}
	blk.Head.TreeHash = blk.CalculateTreeHash()
	headInfo := generateHeadInfo(blk.Head)
	sig, _ := common.Sign(common.Secp256k1, headInfo, acc.Seckey)
	blk.Head.Signature = sig.Encode()

	blockcache.CleanStdVerifier()

	generatedBlockCount.Inc()

	tx.Data.ClearServi(blk.Head.Witness)

	return &blk
}

func generateHeadInfo(head block.BlockHead) []byte {
	var info, numberInfo, versionInfo []byte
	info = make([]byte, 8)
	versionInfo = make([]byte, 4)
	numberInfo = make([]byte, 4)
	binary.BigEndian.PutUint64(info, uint64(head.Time))
	binary.BigEndian.PutUint32(versionInfo, uint32(head.Version))
	binary.BigEndian.PutUint32(numberInfo, uint32(head.Number))
	info = append(info, versionInfo...)
	info = append(info, numberInfo...)
	info = append(info, head.ParentHash...)
	info = append(info, head.TreeHash...)
	info = append(info, head.Info...)
	return common.Sha256(info)
}

func (p *PoB) blockVerify(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
	blockcache.VerifyBlockHead(blk, parent)
	if witnessOfTime(&p.globalStaticProperty, &p.globalDynamicProperty, consensus_common.Timestamp{Slot: blk.Head.Time}) != blk.Head.Witness {
		return nil, errors.New("wrong witness")
	}

	headInfo := generateHeadInfo(blk.Head)
	var signature common.Signature
	signature.Decode(blk.Head.Signature)

	if blk.Head.Witness != common.Base58Encode(signature.Pubkey) {
		return nil, errors.New("wrong pubkey")
	}

	if !common.VerifySignature(headInfo, signature) {
		return nil, errors.New("wrong signature")
	}
	newPool, err := blockcache.StdBlockVerifier(blk, pool)
	if err != nil {
		return nil, err
	}
	return newPool, nil
}
