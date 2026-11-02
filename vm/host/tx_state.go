package host

// CallError returns the first call error in this transaction.
func (h *Host) CallError() error {
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
	h.callErr = nil
	h.settling = false
	h.events = nil
	if h.Rules != nil && h.IsFork4_10_0 {
		h.ClearCacheCost()
	}
}

func (h *Host) checkWrite() error {
	if h.settling {
		return nil
	}
	return h.CallError()
}
