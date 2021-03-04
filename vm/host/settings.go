package host

import "github.com/iost-official/go-iost/v3/core/contract"

// Setting in state db
type Setting struct {
	Costs map[string]contract.Cost `json:"costs"`
}
