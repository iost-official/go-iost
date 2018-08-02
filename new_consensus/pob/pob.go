package pob

import (
	"encoding/binary"

	. "github.com/iost-official/Go-IOS-Protocol/account"
	. "github.com/iost-official/Go-IOS-Protocol/consensus/common"
	. "github.com/iost-official/Go-IOS-Protocol/core/tx"
	. "github.com/iost-official/Go-IOS-Protocol/network"

	"errors"
	"fmt"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/message"

	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/log"
	"github.com/iost-official/Go-IOS-Protocol/vm"
	"github.com/iost-official/Go-IOS-Protocol/vm/lua"
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
	blockChain   block.Chain
	blockCache   blockcache.BlockCache
	router       Router
	synchronizer Synchronizer
	globalStaticProperty
	globalDynamicProperty

	exitSignal chan struct{}
	chBlock    chan message.Message

	log *log.Logger
}

func NewPoB(acc Account, bc block.Chain, pool state.Pool, witnessList []string /*, network core.Network*/) (*PoB, error) {
	//TODO: change initialization based on new interfaces
	p := PoB{
		account: acc,
	}

	p.blockCache = blockcache.NewBlockCache(bc, pool, len(witnessList)*2/3)
	if bc.GetBlockByNumber(0) == nil {

		t := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		err := p.genesis(GetTimestamp(t.Unix()).Slot)
		if err != nil {
			return nil, fmt.Errorf("failed to genesis is nil")
		}
	}

	var err error
	p.router = Route
	if p.router == nil {
		return nil, fmt.Errorf("failed to network.Route is nil")
	}

	p.synchronizer = NewSynchronizer(p.blockCache, p.router, len(witnessList)*2/3)
	if p.synchronizer == nil {
		return nil, err
	}

	p.chBlock, err = p.router.FilteredChan(Filter{
		AcceptType: []ReqType{ReqNewBlock, ReqSyncBlock}})
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

func (p *PoB) initGlobalProperty(acc Account, witnessList []string) {
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

func (p *PoB) genesis(initTime int64) error {

	main := lua.NewMethod(vm.Public, "", 0, 0)

	var code string
	for k, v := range GenesisAccount {
		code += fmt.Sprintf("@PutHM iost %v f%v\n", k, v)
	}

	lc := lua.NewContract(vm.ContractInfo{Prefix: "", GasLimit: 0, Price: 0, Publisher: ""}, code, main)

	tx := Tx{
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
		Content: make([]Tx, 0),
	}
	genesis.Content = append(genesis.Content, tx)
	//TODO: add genesis to db
	return nil
}

func (p *PoB) blockLoop() {
	p.log.I("Start to listen block")
	for {
		select {
		case req, ok := <-p.chBlock:
			if !ok {
				return
			}
			var blk block.Block
			err := blk.Decode(req.Body)
			if err != nil {
				continue
			}
			// TODO
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
		case <-p.exitSignal:
			return
		case <-time.After(time.Second * time.Duration(nextSchedule)):
			currentTimestamp := GetCurrentTimestamp()
			wid := witnessOfTime(&p.globalStaticProperty, &p.globalDynamicProperty, currentTimestamp)
			p.log.I("currentTimestamp: %v, wid: %v, p.account.ID: %v", currentTimestamp, wid, p.account.ID)
			if wid == p.account.ID {
				// TODO
			}
			nextSchedule = timeUntilNextSchedule(&p.globalStaticProperty, &p.globalDynamicProperty, time.Now().Unix())
		}
	}
}

func (p *PoB) genBlock(acc Account, bc block.Chain, pool state.Pool) *block.Block {
	lastBlk := bc.Top()
	blk := block.Block{Content: []Tx{}, Head: block.BlockHead{
		Version:    0,
		ParentHash: lastBlk.HeadHash(),
		Number:     lastBlk.Head.Number + 1,
		Witness:    acc.ID,
		Time:       GetCurrentTimestamp().Slot,
	}}

	vc := vm.NewContext(vm.BaseContext())
	vc.Timestamp = blk.Head.Time
	vc.ParentHash = blk.Head.ParentHash
	vc.BlockHeight = blk.Head.Number
	vc.Witness = vm.IOSTAccount(acc.ID)

	// TODO
	blk.Head.TreeHash = blk.CalculateTreeHash()
	headInfo := generateHeadInfo(blk.Head)
	sig, _ := common.Sign(common.Secp256k1, headInfo, acc.Seckey)
	blk.Head.Signature = sig.Encode()

	blockcache.CleanStdVerifier()

	generatedBlockCount.Inc()

	Data.ClearServi(blk.Head.Witness)

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
	// verify block head
	if err := blockcache.VerifyBlockHead(blk, parent); err != nil {
		return nil, err
	}

	// verify block witness
	if witnessOfTime(&p.globalStaticProperty, &p.globalDynamicProperty, Timestamp{Slot: blk.Head.Time}) != blk.Head.Witness {
		return nil, errors.New("wrong witness")

	}

	headInfo := generateHeadInfo(blk.Head)
	var signature common.Signature
	signature.Decode(blk.Head.Signature)

	if blk.Head.Witness != common.Base58Encode(signature.Pubkey) {
		return nil, errors.New("wrong pubkey")
	}

	// verify block witness signature
	if !common.VerifySignature(headInfo, signature) {
		return nil, errors.New("wrong signature")
	}
	newPool, err := blockcache.StdBlockVerifier(blk, pool)
	if err != nil {
		return nil, err
	}
	return newPool, nil
}
