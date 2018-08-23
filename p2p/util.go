package p2p

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	peer "github.com/libp2p/go-libp2p-peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var (
	errInvalidMultiaddr = errors.New("invalid multiaddr string")
)

// isPortAvailable returns a flag indicating whether or not a TCP port is available.
func isPortAvailable(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
	if err != nil {
		return true
	}
	conn.Close()
	return false
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
