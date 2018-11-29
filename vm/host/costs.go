package host

import "github.com/iost-official/go-iost/core/contract"

// var costs
var (
	Costs = map[string]contract.Cost{
		"PutCost":             contract.NewCost(0, 0, 150),
		"GetCost":             contract.NewCost(0, 0, 100),
		"DelCost":             contract.NewCost(0, 0, 100),
		"KeysCost":            contract.NewCost(0, 0, 100),
		"CompileCost":         contract.NewCost(0, 0, 10),
		"ContextCost":         contract.NewCost(0, 0, 10),
		"DelDelaytxCost":      contract.NewCost(0, 0, 10),
		"DelaytxNotFoundCost": contract.NewCost(0, 0, 10),
		"EventPrice":          contract.NewCost(0, 0, 1),
		"ReceiptPrice":        contract.NewCost(0, 1, 0),
		"OpPrice":             contract.NewCost(0, 0, 1),
		"ErrPrice":            contract.NewCost(0, 0, 1),
	}
)

// EventCost return cost based on event size
func EventCost(size int) contract.Cost {
	return Costs["EventPrice"].Multiply(int64(size))
}

// ReceiptCost based on receipt size
func ReceiptCost(size int) contract.Cost {
	return Costs["ReceiptPrice"].Multiply(int64(size))
}

// CodeSavageCost cost in deploy contract based on code size
func CodeSavageCost(size int) contract.Cost {
	return EventCost(size)
}

// CommonErrorCost returns cost increased by stack layer
func CommonErrorCost(layer int) contract.Cost {
	return Costs["ErrPrice"].Multiply(int64(layer * 10))
}

// CommonOpCost returns cost increased by stack layer
func CommonOpCost(layer int) contract.Cost {
	return Costs["OpPrice"].Multiply(int64(layer * 10))
}
