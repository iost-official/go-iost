package p2p

import (
	"time"

	multiaddr "github.com/multiformats/go-multiaddr"
)

type Config struct {
	SeedNodes   []multiaddr.Multiaddr
	ListenAddr  string
	BucketSize  int
	PeerTimeout time.Duration
	PrivKeyPath string
}
