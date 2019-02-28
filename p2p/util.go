package p2p

import (
	cryrand "crypto/rand"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

var (
	errInvalidMultiaddr = errors.New("invalid multiaddr string")

	ipReg = regexp.MustCompile(`/((25[0-5]|2[0-4]\d|((1\d{2})|([1-9]?\d)))\.){3}(25[0-5]|2[0-4]\d|((1\d{2})|([1-9]?\d)))/`)
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

func getIPFromMaddr(s string) string {
	str := ipReg.FindString(s)
	if len(str) > 2 {
		return str[1 : len(str)-1]
	}
	return ""
}

// private IP:
// 10.0.0.0    - 10.255.255.255
// 192.168.0.0 - 192.168.255.255
// 172.16.0.0  - 172.31.255.255
func privateIP(ip string) (bool, error) {
	IP := net.ParseIP(ip)
	if IP == nil {
		return false, errors.New("invalid IP")
	}
	_, private24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
	_, private20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
	return private24BitBlock.Contains(IP) || private20BitBlock.Contains(IP) || private16BitBlock.Contains(IP), nil
}

func isPublicMaddr(s string) bool {
	ip := getIPFromMaddr(s)
	if ip == "127.0.0.1" {
		return false
	}
	private, err := privateIP(ip)
	if err != nil {
		return false
	}
	return !private
}

func randomPID() (peer.ID, error) {
	_, pubkey, err := crypto.GenerateEd25519Key(cryrand.Reader)
	if err != nil {
		return "", err
	}
	return peer.IDFromPublicKey(pubkey)
}
