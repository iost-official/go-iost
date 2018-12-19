package database

const (
	delaytxPrefix = "t-"
)

// DelaytxHandler handler of delay tx
type DelaytxHandler struct {
	db database
}

func (m *DelaytxHandler) delaytxKey(txHash string) string {
	return delaytxPrefix + txHash
}

// StoreDelaytx stores delaytx hash.
func (m *DelaytxHandler) StoreDelaytx(txHash, publisher string) {
	m.db.Put(m.delaytxKey(txHash), publisher)
}

// GetDelaytx gets the delay tx's publisher.
func (m *DelaytxHandler) GetDelaytx(txHash string) string {
	return m.db.Get(m.delaytxKey(txHash))
}

// HasDelaytx checks whether the delaytx exists.
func (m *DelaytxHandler) HasDelaytx(txHash string) bool {
	return m.db.Has(m.delaytxKey(txHash))
}

// DelDelaytx deletes the delaytx hash.
func (m *DelaytxHandler) DelDelaytx(txHash string) {
	m.db.Del(m.delaytxKey(txHash))
}
