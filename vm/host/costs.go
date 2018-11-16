package host

import "github.com/iost-official/go-iost/core/contract"

// var list cost
var (
	PutCost  = contract.NewCost(0, 0, 12)
	GetCost  = contract.NewCost(0, 0, 8)
	DelCost  = contract.NewCost(0, 0, 8)
	KeysCost = contract.NewCost(0, 0, 12)

	CompileErrCost       = contract.NewCost(0, 0, 10)
	ContractNotFoundCost = contract.NewCost(0, 0, 10)
	ABINotFoundCost      = contract.NewCost(0, 0, 11)

	DelContractCost = contract.NewCost(0, 0, 10)

	BlockInfoCost   = contract.NewCost(0, 0, 1)
	TxInfoCost      = contract.NewCost(0, 0, 1)
	ContextInfoCost = contract.NewCost(0, 0, 1)

	TransferCost = contract.NewCost(0, 0, 3)

	RequireAuthCost = contract.NewCost(0, 0, 1)

	PledgeForGasCost = contract.NewCost(0, 0, 20)

	DelDelaytxCost      = contract.NewCost(0, 0, 10)
	DelaytxNotFoundCost = contract.NewCost(0, 0, 10)
)

// EventCost return cost based on event size
func EventCost(size int) contract.Cost {
	return contract.NewCost(0, int64(size), 1)
}

// ReceiptCost based on receipt size
func ReceiptCost(size int) contract.Cost {
	return EventCost(size)
}

// CodeSavageCost cost in deploy contract based on code size
func CodeSavageCost(size int) contract.Cost {
	return EventCost(size)
}

// CommonErrorCost returns cost increased by stack layer
func CommonErrorCost(layer int) contract.Cost {
	return contract.NewCost(0, 0, int64(layer*10))
}

// CommonOpCost returns cost increased by stack layer
func CommonOpCost(layer int) contract.Cost {
	return contract.NewCost(0, 0, int64(layer*10))
}
