package dpos

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	. "github.com/iost-official/prototype/account"
	. "github.com/iost-official/prototype/consensus/common"
	. "github.com/iost-official/prototype/core/tx"
	. "github.com/iost-official/prototype/network"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/block"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/core/state"
	"errors"
)

type DPoS struct {
	account      Account
	blockCache   BlockCache
	router       Router
	synchronizer Synchronizer
	globalStaticProperty
	globalDynamicProperty

	//测试用，保存投票状态，以及投票消息内容的缓存
	votedStats map[string][]string
	infoCache  [][]byte

	exitSignal chan bool
	chTx       chan message.Message
	chBlock    chan message.Message
}

// NewDPoS: 新建一个DPoS实例
// acc: 节点的Coinbase账户, bc: 基础链(从数据库读取), witnessList: 见证节点列表
func NewDPoS(acc Account, bc block.Chain, pool state.Pool, witnessList []string /*, network core.Network*/) (*DPoS, error) {
	p := DPoS{}
	p.Account = acc
	p.blockCache = NewBlockCache(bc, pool, len(witnessList)*2/3+1)

	var err error
	p.router, err = RouterFactory("base")
	if err != nil {
		return nil, err
	}

	p.synchronizer = NewSynchronizer(p.blockCache, p.router)
	if p.synchronizer == nil {
		return nil, err
	}

	//	Tx chan init
	p.chTx, err = p.router.FilteredChan(Filter{
		WhiteList:  []message.Message{},
		BlackList:  []message.Message{},
		RejectType: []ReqType{},
		AcceptType: []ReqType{
			ReqPublishTx,
			reqTypeVoteTest, // Only for test
		}})
	if err != nil {
		return nil, err
	}

	//	Block chan init
	p.chBlock, err = p.router.FilteredChan(Filter{
		WhiteList:  []message.Message{},
		BlackList:  []message.Message{},
		RejectType: []ReqType{},
		AcceptType: []ReqType{ReqNewBlock}})
	if err != nil {
		return nil, err
	}

	p.initGlobalProperty(p.Account, witnessList)
	return &p, nil
}

func (p *DPoS) initGlobalProperty(acc Account, witnessList []string) {
	p.globalStaticProperty = newGlobalStaticProperty(acc, witnessList)
	p.globalDynamicProperty = newGlobalDynamicProperty()
}

// Run: 运行DPoS实例
func (p *DPoS) Run() {
	//go p.blockLoop()
	//go p.scheduleLoop()
	p.genBlock(p.Account, block.Block{})
}

// Stop: 停止DPoS实例
func (p *DPoS) Stop() {
	close(p.chTx)
	close(p.chBlock)
	p.exitSignal <- true
}

func (p *DPoS) genesis(initTime int64) error {
	genesis := &block.Block{
		Head: block.BlockHead{
			Version: 0,
			Number:  0,
			Time:    initTime,
		},
		Content: make([]Tx, 0),
	}
	p.blockCache.AddGenesis(genesis)
	return nil
}

func (p *DPoS) txListenLoop() {
	for {
		req, ok := <-p.chTx
		if !ok {
			return
		}
		if req.ReqType == reqTypeVoteTest {
			p.addWitnessMsg(req)
			continue
		}
		var tx Tx
		tx.Decode(req.Body)
		p.router.Send(req)
		if VerifyTxSig(tx) {
			p.blockCache.AddTx(&tx)
		}
	}
}

func (p *DPoS) blockLoop() {
	//收到新块，验证新块，如果验证成功，更新DPoS全局动态属性类并将其加入block cache，再广播
	verifier := func(blk *block.Block, parent *block.Block, pool state.Pool) (state.Pool, error) {
		// verify block head

		if err := VerifyBlockHead(blk, parent); err != nil {
			return nil, err
		}

		// verify block witness
		if witnessOfTime(&p.globalStaticProperty, &p.globalDynamicProperty, Timestamp{blk.Head.Time}) != blk.Head.Witness {
			return nil, errors.New("wrong witness")
		}

		headInfo := generateHeadInfo(blk.Head)
		var signature common.Signature
		signature.Decode(blk.Head.Signature)

		// verify block witness signature
		if !common.VerifySignature(headInfo, signature) {
			return nil, errors.New("wrong signature")
		}
		newPool, err := StdBlockVerifier(blk, pool)
		if err != nil {
			return nil, err
		}
		return newPool, nil
	}

	for {
		req, ok := <-p.chBlock
		if !ok {
			return
		}
		var blk block.Block
		blk.Decode(req.Body)
		err := p.blockCache.Add(&blk, verifier)
		if err != ErrBlock {
			p.globalDynamicProperty.update(&blk.Head)
			if err == ErrNotFound {
				// New block is a single block
				need, start, end := p.synchronizer.NeedSync(uint64(blk.Head.Number))
				if need {
					go p.synchronizer.SyncBlocks(start, end)
				}
			}
		}
		ts := Timestamp{blk.Head.Time}
		if ts.After(p.globalDynamicProperty.NextMaintenanceTime) {
			p.performMaintenance()
		}
	}
}

func (p *DPoS) scheduleLoop() {
	//通过时间判定是否是本节点的slot，如果是，调用产生块的函数，如果不是，设定一定长的timer睡眠一段时间

	for {
		currentTimestamp := GetCurrentTimestamp()
		wid := witnessOfTime(&p.globalStaticProperty, &p.globalDynamicProperty, currentTimestamp)
		if wid == p.Account.ID {
			bc := p.blockCache.LongestChain()
			blk := p.genBlock(p.Account, *bc.Top())
			p.router.Send(message.Message{Body: blk.Encode()}) //??
		}
		nextSchedule := timeUntilNextSchedule(&p.globalStaticProperty, &p.globalDynamicProperty, time.Now().Unix())
		time.Sleep(time.Duration(nextSchedule))
	}
}

func (p *DPoS) genBlock(acc Account, lastBlk block.Block) *block.Block {
	blk := block.Block{Content: []Tx{}, Head: block.BlockHead{
		Version:    0,
		ParentHash: lastBlk.Head.BlockHash,
		TreeHash:   make([]byte, 0),
		BlockHash:  make([]byte, 0),
		Info:       encodeDPoSInfo(p.infoCache),
		Number:     lastBlk.Head.Number + 1,
		Witness:    acc.ID,
		Time:       GetCurrentTimestamp().Slot,
	}}
	p.infoCache = [][]byte{}
	headInfo := generateHeadInfo(blk.Head)
	fmt.Println(acc.Seckey)
	sig, _ := common.Sign(common.Secp256k1, headInfo, acc.Seckey)
	blk.Head.Signature = sig.Encode()
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
	return info
}

// 测试函数，用来将info和vote消息进行转换，在块被确认时被调用
// TODO:找到适当的时机调用
func decodeDPoSInfo(info []byte) [][]byte {
	votes := bytes.Split(info, []byte("/"))
	return votes
}

// 测试函数，用来将info和vote消息进行转换，在生成块的时候调用
func encodeDPoSInfo(votes [][]byte) []byte {
	var info []byte
	for _, req := range votes {
		info = append(info, req...)
		info = append(info, byte('/'))
	}
	return info
}
