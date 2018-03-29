package discover

import (
	"net"
	"common"
	"time"
	"crypto"
)

type NodeID [NodeIDBits / 8]byte

const NodeIDBits = 512

type Node struct {
	IP       net.IP // len 4 for IPv4 or 16 for IPv6
	UDP, TCP uint16 // port numbers
	ID       NodeID // the node's public key

	// This is a cached copy of sha3(ID) which is used for node
	// distance calculations. This is part of Node in order to make it
	// possible to write tests that need a node at a certain distance.
	// In those tests, the content of sha will not actually correspond
	// with ID.
	sha common.Hash

	// Time when the node was added to the table.
	addedAt time.Time
}

// NewNode creates a new node. It is mostly meant to be used for
// testing purposes.
func NewNode(id NodeID, ip net.IP, udpPort, tcpPort uint16) *Node {
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	return &Node{
		IP:  ip,
		UDP: udpPort,
		TCP: tcpPort,
		ID:  id,
		sha: crypto.Keccak256Hash(id[:]),
		//TODO: implement Keccak256
	}
}
