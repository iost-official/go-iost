package p2p

import (
	"errors"
	"fmt"
	"net"
	"strings"

	peer "github.com/libp2p/go-libp2p-peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var (
	errInvalidMultiaddr = errors.New("invalid multiaddr string")
)

// isPortAvailable returns a flag indicating whether or not a TCP port is available.
func isPortAvailable(port int) bool {
	conn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func parseMultiaddr(s string) (peer.ID, multiaddr.Multiaddr, error) {
	strs := strings.Split(s, "/ipfs/")
	if len(strs) != 2 {
		return "", nil, errInvalidMultiaddr
	}
	addr, err := multiaddr.NewMultiaddr(strs[0])
	if err != nil {
		return "", nil, err
	}
	peerID, err := peer.IDB58Decode(strs[1])
	if err != nil {
		return "", nil, err
	}
	return peerID, addr, nil
}
