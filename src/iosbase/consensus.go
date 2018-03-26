package iosbase

type TxStatus int

const (
	ACCEPT TxStatus = iota // 交易已被记录
	CACHED                 // 交易在一个分叉上，有可能被删除
	POOL                   // 交易在交易池里，还没被记录
	REJECT                 // 交易有冲突，无效
)

type Consensus interface {
	Init(bc *BlockChain, sp *StatePool, network Network) error // 依赖注入，由共识机制来管理区块链和UTXO池，但是选用区块链、utxo池的哪种实现取决于外部
	Run()
	Stop()

	PublishTx(tx Tx) error           // 向共识机制发布一条交易
	CheckTx(tx Tx) (TxStatus, error) // 询问一个交易的状态

	GetStatus() (BlockChain, StatePool, error)       // 不包含分叉的、已经确认不会被更改的区块链部分，以及对应的UTXO Pool
	GetCachedStatus() (BlockChain, StatePool, error) // 包含分叉、可能会被修改的区块链
}
