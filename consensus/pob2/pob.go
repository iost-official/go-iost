package pob2

import (
	"bytes"
	"encoding/binary"

	. "github.com/iost-official/prototype/account"
	. "github.com/iost-official/prototype/consensus/common"
	. "github.com/iost-official/prototype/core/tx"
	. "github.com/iost-official/prototype/network"

	"errors"
	"fmt"
	"time"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/blockcache"
	"github.com/iost-official/prototype/core/message"

	"github.com/iost-official/prototype/core/state"
	"github.com/iost-official/prototype/core/txpool"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/verifier"
	"github.com/iost-official/prototype/vm"
	"github.com/iost-official/prototype/vm/lua"
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

var TxPerBlk int

type PoB struct {
	account      Account
	blockCache   blockcache.BlockCache
	router       Router
	synchronizer Synchronizer
	globalStaticProperty
	globalDynamicProperty

	//测试用，保存投票状态，以及投票消息内容的缓存
	votedStats map[string][]string
	infoCache  [][]byte

	exitSignal chan struct{}
	chBlock    chan message.Message

	log *log.Logger
}

// NewPoB: 新建一个PoB实例
// acc: 节点的Coinbase账户, bc: 基础链(从数据库读取), pool: 基础state池（从数据库读取）, witnessList: 见证节点列表
func NewPoB(acc Account, bc block.Chain, pool state.Pool, witnessList []string /*, network core.Network*/) (*PoB, error) {
	TxPerBlk = 800
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

	//	Block chan init
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

// Run: 运行PoB实例
func (p *PoB) Run() {
	p.synchronizer.StartListen()
	go p.blockLoop()
	go p.scheduleLoop()
	//p.genBlock(p.Account, block.Block{})
}

// Stop: 停止PoB实例
func (p *PoB) Stop() {
	close(p.chBlock)
	close(p.exitSignal)
}

func (p *PoB) BlockCache() blockcache.BlockCache {
	return p.blockCache
}

// BlockChain 返回已确认的block chain
func (p *PoB) BlockChain() block.Chain {
	return p.blockCache.BlockChain()
}

// CachedBlockChain 返回缓存中的最长block chain
func (p *PoB) CachedBlockChain() block.Chain {
	return p.blockCache.LongestChain()
}

// StatePool 返回已确认的state pool
func (p *PoB) StatePool() state.Pool {
	return p.blockCache.BasePool()
}

// CacheStatePool 返回缓存中最新的state pool
func (p *PoB) CachedStatePool() state.Pool {
	return p.blockCache.LongestPool()
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
	stp, err := verifier.ParseGenesis(tx.Contract, p.StatePool())
	if err != nil {
		panic("failed to ParseGenesis")
	}

	err = p.blockCache.SetBasePool(stp)
	if err != nil {
		panic("failed to SetBasePool")
	}

	err = p.blockCache.AddGenesis(genesis)
	if err != nil {
		panic("failed to AddGenesis")
	}
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

			p.log.I("Received block:%v ,from=%v, timestamp: %v, Witness: %v, trNum: %v", blk.Head.Number, req.From, blk.Head.Time, blk.Head.Witness, len(blk.Content))
			localLength := p.blockCache.ConfirmedLength()
			if blk.Head.Number > int64(localLength)+MaxAcceptableLength {
				// Do not accept block of too height, must wait for synchronization
				if req.ReqType == int32(ReqNewBlock) {
					go p.synchronizer.SyncBlocks(localLength, localLength+uint64(MaxAcceptableLength))
				}
				continue
			}
			err = p.blockCache.Add(&blk, p.blockVerify)
			if err == nil {
				p.log.I("Link it onto cached chain")
				p.blockCache.SendOnBlock(&blk)
				receivedBlockCount.Inc()
			} else {
				p.log.I("Error: %v", err)
				//p.log.I("[blockloop]:verify blk faild\n%s\n", &blk)
			}
			if err != blockcache.ErrBlock && err != blockcache.ErrTooOld {
				go p.synchronizer.BlockConfirmed(blk.Head.Number)
				if err == nil {
					p.globalDynamicProperty.update(&blk.Head)
				} else if err == blockcache.ErrNotFound && req.ReqType == int32(ReqNewBlock) {
					// New block is a single block
					need, start, end := p.synchronizer.NeedSync(uint64(blk.Head.Number))
					if need {
						go p.synchronizer.SyncBlocks(start, end)
					}
				}
			}
			/*
								ts := Timestamp{blk.Head.Time}
								if ts.After(p.globalDynamicProperty.NextMaintenanceTime) {
									p.performMaintenance()
				 				}
			*/
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) scheduleLoop() {
	//通过时间判定是否是本节点的slot，如果是，调用产生块的函数，如果不是，设定一定长的timer睡眠一段时间
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

				//todo test
				bc := p.blockCache.LongestChain()
				iter := bc.Iterator()
				for {
					block := iter.Next()
					if block == nil {
						break
					}
					confirmedBlockchainLength.Set(float64(p.blockCache.ConfirmedLength()))
					p.log.I("CBC ConfirmedLength: %v, block Number: %v, witness: %v", p.blockCache.ConfirmedLength(), block.Head.Number, block.Head.Witness)
				}
				// end test

				// TODO 考虑更好的解决方法，因为两次调用之间可能会进入新块影响最长链选择

				pool := p.blockCache.LongestPool()
				blk := p.genBlock(p.account, bc, pool)

				p.globalDynamicProperty.update(&blk.Head)
				p.log.I("Generating block, current timestamp: %v number: %v", currentTimestamp, blk.Head.Number)

				bb := blk.Encode()
				msg := message.Message{ReqType: int32(ReqNewBlock), Body: bb}
				log.Log.I("Block size: %v, TrNum: %v", len(bb), len(blk.Content))
				go p.router.Broadcast(msg)
				p.chBlock <- msg
				p.log.I("Broadcasted block, current timestamp: %v number: %v", currentTimestamp, blk.Head.Number)
			}
			nextSchedule = timeUntilNextSchedule(&p.globalStaticProperty, &p.globalDynamicProperty, time.Now().Unix())
		}
	}
}

func (p *PoB) genBlock(acc Account, bc block.Chain, pool state.Pool) *block.Block {
	limitTime := time.NewTicker(((SlotLength/3 - 1) + 1) * time.Second)
	lastBlk := bc.Top()
	blk := block.Block{Content: []Tx{}, Head: block.BlockHead{
		Version:    0,
		ParentHash: lastBlk.HeadHash(),
		Info:       encodePoBInfo(p.infoCache),
		Number:     lastBlk.Head.Number + 1,
		Witness:    acc.ID,
		Time:       GetCurrentTimestamp().Slot,
	}}
	p.infoCache = [][]byte{}
	//return &blk
	spool1 := pool.Copy()

	vc := vm.NewContext(vm.BaseContext())
	vc.Timestamp = blk.Head.Time
	vc.ParentHash = blk.Head.ParentHash
	vc.BlockHeight = blk.Head.Number
	vc.Witness = vm.IOSTAccount(acc.ID)

	//TODO Content大小控制

	txCnt := TxPerBlk + rand.Intn(500)
	var tx TransactionsList
	if txpool.TxPoolS != nil {
		p.log.I("PendingTransactions Begin...")
		tx = txpool.TxPoolS.PendingTransactions(txCnt)
		p.log.I("PendingTransactions End.")
		txPoolSize.Set(float64(txpool.TxPoolS.TransactionNum()))
		p.log.I("PendingTransactions Size: %v.", txpool.TxPoolS.PendingTransactionNum())
	}

	if len(tx) != 0 {
	ForEnd:
		for _, t := range tx {
			select {
			case <-limitTime.C:
				p.log.I("Gen Block Time Limit.")
				break ForEnd
			default:
				if len(blk.Content) >= txCnt {
					p.log.I("Gen Block Tx Number Limit.")
					break ForEnd
				}
				if err := blockcache.StdCacheVerifier(t, spool1, vc); err == nil {
					blk.Content = append(blk.Content, *t)

				}
			}
		}
	}
	blk.Head.TreeHash = blk.CalculateTreeHash()
	headInfo := generateHeadInfo(blk.Head)
	sig, _ := common.Sign(common.Secp256k1, headInfo, acc.Seckey)
	blk.Head.Signature = sig.Encode()

	blockcache.CleanStdVerifier() // hpj: 现在需要手动清理缓存的虚拟机

	generatedBlockCount.Inc()

	//Clear Servi
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

// 测试函数，用来将info和vote消息进行转换，在块被确认时被调用
// TODO:找到适当的时机调用
func decodePoBInfo(info []byte) [][]byte {
	votes := bytes.Split(info, []byte("/"))
	return votes
}

// 测试函数，用来将info和vote消息进行转换，在生成块的时候调用
func encodePoBInfo(votes [][]byte) []byte {
	var info []byte
	for _, req := range votes {
		info = append(info, req...)
		info = append(info, byte('/'))
	}
	return info
}

//收到新块，验证新块，如果验证成功，更新PoB全局动态属性类并将其加入block cache，再广播
func (p *PoB) blockVerify(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
	// verify block head

	if err := blockcache.VerifyBlockHead(blk, parent); err != nil {
		return nil, err
	}

	// verify block witness
	// TODO currentSlot is negative
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
