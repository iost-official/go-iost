package dpos

import (
	"github.com/iost-official/prototype/core"
	. "github.com/iost-official/prototype/p2p"
	. "github.com/iost-official/prototype/pow"
)

type DPoS struct {
	core.Member
	BlockCache
	Router
	GlobalStaticProperty
	GlobalDynamicProperty

	//测试用，保存投票状态，以及投票消息的缓存
	votedStats map[string][]string
	infoCache  []core.Request

	ExitSignal chan bool
	chTx       chan core.Request
	chBlock    chan core.Request
}

func NewDPoS(mb core.Member, bc core.BlockChain, network core.Network) (*DPoS, error) {
	// Member初始化
	p.Member = mb
	p := DPoS{}
	p.BlockCache = NewBlockCache(bc, 6)

	p.Router, err = RouterFactor()
	if err != nil {
		return err
	}
	p.Router.Init(network)

	/*
		chan init
	*/
	p.initGlobalProperty(p.Member, make([]string))
	return &p, nil
}

func (p *DPoS) initGlobalProperty(mb core.Member, witnessList []string) {
	p.GlobalStaticProperty = NewGlobalStaticProperty(mb, witnessList)
	p.GlobalDynamicProperty = NewGlobalDynamicProperty()
}

func (p *DPoS) Run() {
	go p.blockLoop()
	go p.scheduleLoop()
}

func (p *DPoS) Stop() {
	close(p.chTx)
	close(p.chBlock)
	p.ExitSignal <- true
}

func (p *DPoS) Genesis(initTime Timestamp, hash []byte) error {
	return nil
}

func (p *DPos) txListeniLoop() {
	for {
		req, ok := <-p.chTx
		if !ok {
			return
		}
		var tx core.Tx
		tx.Decode(req.Body)
		p.PublishTx(tx)
	}
}

func (p *DPoS) blockLoop() {
	//收到新块，验证新块，如果验证成功，更新DPoS全局动态属性类并将其加入block cache，再广播
	for {
		req, ok := <-p.chBlock
		if !ok {
			return
		}
		var blk core.Block
		blk.Decode(req.Body)
		p.BlockCache.Add(&blk)
	}
}

func (p *DPoS) scheduleLoop() {
	//通过时间判定是否是本节点的slot，如果是，调用产生块的函数，如果不是，设定一定长的timer睡眠一段时间
	for {
		currentTimestamp := GetCurrentTimestamp()
		wid := WitnessOfTime(p.GlobalStaticProperty, p.GlobalDynamicProperty, currentTimestamp)
		if wid == p.Member.ID {
			//TODO
			// 生成blk
		}

	}
}
