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

func TestGetIPFromMa(t *testing.T) {
	for _, testCase := range []struct {
		expect string
		input  string
	}{
		{"127.0.0.1", "/ip4/127.0.0.1/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM"},
		{"172.17.0.2", "/ip4/172.17.0.2/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM"},
		{"56.12.32.32", "/ip4/56.12.32.32/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM"},
		{"", "/ip4/256.12.32.32/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM"},
		{"", "/ip4/56.12.32.321/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM"},
	} {
		assert.Equal(t, testCase.expect, getIPFromMaddr(testCase.input))
	}
}

func TestPrivateIP(t *testing.T) {
	for _, testCase := range []struct {
		input  string
		expect bool
	}{
		{"10.32.1.3", true},
		{"192.168.3.2", true},
		{"192.167.3.2", false},
		{"172.16.0.1", true},
		{"172.15.31.32", false},
		{"127.0.0.1", false},
	} {
		b, err := privateIP(testCase.input)
		assert.Nil(t, err)
		assert.Equal(t, testCase.expect, b)
	}
}

func TestIsPublicMa(t *testing.T) {
	for _, testCase := range []struct {
		input  string
		expect bool
	}{
		{"/ip4/127.0.0.1/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM", false},
		{"/ip4/172.17.0.2/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM", false},
		{"/ip4/10.3.2.1/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM", false},
		{"/ip4/56.12.32.32/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM", true},
		{"/ip4/256.12.32.32/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM", false},
		{"/ip4/56.12.32.321/tcp/1111/ipfs/Qmb6ib8i3B95HuGRoC2KTy5dzxeP4LLYQkxPUiGFiiiUtM", false},
	} {
		assert.Equal(t, testCase.expect, isPublicMaddr(testCase.input))
	}
}
