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

func NewDPoS() (*DPoS, error) {
	p := DPoS{}
	p.BlockCache = NewBlockCache()
	p.Router = RouterFactor()
	p.InitGlobalProperty()
	return &p, nil
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

func (p *DPoS) initGlobalProperty(id string, witnessList []string) {
	p.GlobalStaticProperty = NewGlobalStaticProperty(id, witnessList)
	p.GlobalDynamicProperty = NewGlobalDynamicProperty()
}

func (p *DPoS) Add(block *core.Block) {
	p.BlockCache.Add(block)
}

func (p *DPoS) Genesis(initTime Timestamp, hash []byte) error {
}

func (p *DPoS) blockLoop() {
	//收到新块，验证新块，如果验证成功，更新DPoS全局动态属性类并将其加入block cache，再广播
}

func (p *DPoS) scheduleLoop() {
	//通过时间判定是否是本节点的slot，如果是，调用产生块的函数，如果不是，设定一定长的timer睡眠一段时间

}
