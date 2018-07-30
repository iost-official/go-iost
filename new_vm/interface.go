package new_vm

import "github.com/iost-official/Go-IOS-Protocol/common"

// define in other module

type Receipt struct {

}

// Transaction 的实现
type Tx struct {
	Time      int64
	Nonce     int64
	// Contract  vm.Contract
	Actions   []Action
	Signs     []common.Signature
	Publisher common.Signature
	Recorder  common.Signature
}


// Action 的实现
type Action struct {
	Contract    string  // 为空则视为调用系统合约
	ActionName  string
	Data        string  // json
}


// Transaction Receipt 实现
type TxReceipt struct {
	TxHash      []byte
	GasUsage    uint64
	// 目前只收gas，这些可以先没有
	/*
	CpuTimeUsage    uint64
	NetUsage    uint64
	RAMUsage    uint64
	*/
	Status      Status
	Receipts    []Receipt
}


type Code int
const (
	Success Code = iota
	ErrorGasRunOut
	ErrorBalanceNotEnough
	ErrorOverFlow
	ErrorParamter
	ErrorUnknown
)
type Status struct {
	Code    Code
	Message string
}


// end

type Engine interface {
	Exec(tx0 Tx)
}
