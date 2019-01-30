package database

import (
	"strings"
)

const (
	delaytxPrefix = "t-"

	deferSep = "@"
)

// DelaytxHandler handler of delay tx
type DelaytxHandler struct {
	db database
}

func (m *DelaytxHandler) delaytxKey(txHash string) string {
	return delaytxPrefix + txHash
}

// StoreDelaytx stores delaytx hash.
func (m *DelaytxHandler) StoreDelaytx(txHash, publisher, deferTxHash string) {
	m.db.Put(m.delaytxKey(txHash), publisher+deferSep+deferTxHash)
}

// GetDelaytx gets the delay tx's publisher and deferTxHash.
func (m *DelaytxHandler) GetDelaytx(txHash string) (string, string) {
	str := m.db.Get(m.delaytxKey(txHash))
	if str == NilPrefix {
		return "", ""
	}
	arr := strings.SplitN(str, deferSep, 2)
	if len(arr) != 2 {
		return "", ""
	}
	return arr[0], arr[1]
}

// HasDelaytx checks whether the delaytx exists.
func (m *DelaytxHandler) HasDelaytx(txHash string) bool {
	return m.db.Has(m.delaytxKey(txHash))
}

// DelDelaytx deletes the delaytx hash.
func (m *DelaytxHandler) DelDelaytx(txHash string) {
	m.db.Del(m.delaytxKey(txHash))
}
