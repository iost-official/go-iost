package p2p

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPortAvailable(t *testing.T) {
	assert.True(t, isPortAvailable(8888))
	l, err := net.Listen("tcp", "127.0.0.1:8888")
	assert.Nil(t, err)
	assert.False(t, isPortAvailable(8888))
	assert.Nil(t, l.Close())
}

func TestParseMultiaddr(t *testing.T) {
	peerID, addr, err := parseMultiaddr("/ip4/127.0.0.1/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM")
	assert.Nil(t, err)
	assert.Equal(t, "Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM", peerID.Pretty())
	assert.Equal(t, "/ip4/127.0.0.1/tcp/1111", addr.String())
}
