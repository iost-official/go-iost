package drp

type Client interface {
	RequestRandom(seed []byte) []byte
	AddServer(server Server)
	OnReceivingProof()
	OnReceivingSignOff()
}
