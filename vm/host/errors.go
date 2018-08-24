package host

import "errors"

// var ...
var (
	ErrBalanceNotEnough = errors.New("balance not enough")
	ErrTransferNegValue = errors.New("trasfer amount less than zero")
	ErrReenter          = errors.New("re-entering")
	ErrPermissionLost   = errors.New("transaction has no permission")

	ErrContractNotFound = errors.New("contract not exists")
	ErrUpdateRefused    = errors.New("update refused")
	ErrDestroyRefused   = errors.New("destroy refused")
)
