package tx

type txCache interface {
	PushTx() error
	GetTx(hash []byte) error
}
