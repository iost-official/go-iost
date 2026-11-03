package host

import (
	"errors"
	"fmt"
)

// ErrContractChangeRejected identifies a local deployment or update rejection.
var ErrContractChangeRejected = errors.New("contract deployment and update temporarily disabled")

// SetContractPolicy selects local restrictions before transaction execution.
func (h *Host) SetContractPolicy(enabled bool) {
	h.restrictContracts = enabled
}

// ContractPolicy reports whether the pre-activation restriction is active.
func (h *Host) ContractPolicy() bool {
	return h.restrictContracts && h.Rules != nil && !h.IsFork4_10_0
}

// CheckContractChange records a restricted native entry until transaction cleanup.
func (h *Host) CheckContractChange(api string) error {
	if h.ContractPolicy() && h.contractPolicyErr == nil {
		h.contractPolicyErr = fmt.Errorf("%w: system.iost/%s", ErrContractChangeRejected, api)
	}
	return h.contractPolicyErr
}

// ContractPolicyError returns the transaction's local rejection, if any.
func (h *Host) ContractPolicyError() error {
	return h.contractPolicyErr
}
