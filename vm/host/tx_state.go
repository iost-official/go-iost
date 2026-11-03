package host

// CallError returns a policy rejection or the first active call error.
func (h *Host) CallError() error {
	if h.contractPolicyErr != nil {
		return h.contractPolicyErr
	}
	if h.Rules == nil || !h.IsFork4_10_0 {
		return nil
	}
	return h.callErr
}

// SetCallError keeps a call error until the transaction is cleared.
func (h *Host) SetCallError(err error) {
	if h.Rules != nil && h.IsFork4_10_0 && h.callErr == nil {
		h.callErr = err
	}
}

// ResetTxState clears transaction-local execution state.
func (h *Host) ResetTxState() {
	if h.contractPolicyErr != nil || (h.Rules != nil && h.IsFork4_10_0) {
		h.ClearCacheCost()
	}
	h.contractPolicyErr = nil
	h.callErr = nil
	h.settling = false
	h.events = nil
}

func (h *Host) checkWrite() error {
	if h.settling {
		return nil
	}
	return h.CallError()
}
