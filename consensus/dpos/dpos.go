package dpos

import (
	"github.com/iost-official/prototype/core"
	. "github.com/iost-official/prototype/p2p"
	"sync"
)

type DPoS struct {
	core.Member
	Recorder //what is this?
	BlockCacheImpl
	p2p.Router
	GlobalStaticProperty
	GlobalDynamicProperty

	blockUpdateLock sync.RWMutex
	ExitSignal chan bool
	chTx       chan core.Request
	chBlock    chan core.Request
}

func NewDPoS() *DPoS {
	p := DPoS{}
	p.Init()
	return &p
}

func (p *DPoS) Init() {
	p.BlockCacheImpl = NewBlockCache()
	p.Router = RouterFactor()

	p.InitGlobalProperty()
}

func (p *DPoS) Run() {

}

func (p *DPoS) InitGlobalProperty(id string, witnessList []string) {
	p.GlobalStaticProperty = NewGlobalStaticProperty(id, witnessList)
	p.GlobalDynamicProperty = NewGlobalDynamicProperty()
}

func (p *DPoS) Add(block *core.Block) {
	p.BlockCacheImpl.Add(block)
}

func (p *DPoS) blockLoop() {
	//收到新块，验证新块，如果验证成功，更新DPoS全局动态属性类并将其加入block cache，再广播
}

func (p *DPoS) scheduleLoop() {
	//通过时间判定是否是本节点的slot，如果是，调用产生块的函数，如果不是，设定一定长的timer睡眠一段时间

}
