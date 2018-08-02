package tx

/**
 * Describtion: tx
 * User: wangyu
 * Date: 18-7-30
 */

type StatusCode int
// tx execution result
const (
	Success StatusCode = iota
	ErrorGasRunOut
	ErrorBalanceNotEnough
	ErrorOverFlow			// number calculation over flow
	ErrorParamter			// paramter mismatch when calling function
	ErrorUnknown			// other errors
)
type Status struct {
	Code    StatusCode
	Message string
}

type ReceiptType int

const (
	SystemDefined ReceiptType = iota
	UserDefined
)

// Receipt generated when applying transaction
type Receipt struct {
	Type 		ReceiptType			// system defined or user defined receipt type
	Content     string				// can be a raw string or a json string
}

// TxReceipt Transaction Receipt 实现
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
	SuccActionNum	int		// 执行成功的 action 个数
	Receipts    []Receipt
}s
