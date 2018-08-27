package host

import "github.com/iost-official/Go-IOS-Protocol/core/contract"

var (
	PutCost  = contract.NewCost(100, 0, 1)
	GetCost  = contract.NewCost(100, 0, 1)
	DelCost  = contract.NewCost(1, 0, 1)
	KeysCost = contract.NewCost(100, 0, 1)

	CompileErrCost       = contract.NewCost(0, 0, 10)
	ContractNotFoundCost = contract.NewCost(0, 0, 10)
	ABINotFoundCost      = contract.NewCost(0, 0, 11)

	DelContractCost = contract.NewCost(0, 0, 10)

	BlockInfoCost = contract.NewCost(0, 0, 1)
	TxInfoCost    = contract.NewCost(0, 0, 1)

	TransferCost = contract.NewCost(300, 0, 3)

	RequireAuthCost = contract.NewCost(0, 0, 1)
)

func EventCost(size int) *contract.Cost {
	return contract.NewCost(1, int64(size/100), 1)
}

func ReceiptCost(size int) *contract.Cost {
	return EventCost(size)
}

func CodeSavageCost(size int) *contract.Cost {
	return EventCost(size)
}

func CommonErrorCost(layer int) *contract.Cost {
	return contract.NewCost(0, 0, int64(layer*10))
}
