package discover

import (
	"errors"
	"net"
	//	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/iost-official/prototype/common"
	//	"github.com/iost-official/prototype/params"
	//	"math/rand"
)

// NodeID is a node's identity.
type NodeID string

// Node represents a connected remote node.
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

// NewNode returns a new Node instance.
func NewNode(id NodeID, ip net.IP, udpPort, tcpPort uint16) *Node {
	if ipv4 := ip.To4(); ipv4 != nil {
		ip = ipv4
	}
	return &Node{
		IP:  ip,
		UDP: udpPort,
		TCP: tcpPort,
		ID:  id,
	}
}

// Incomplete returns true for nodes with no IP address.
func (n *Node) Incomplete() bool {
	return n.IP == nil
}

// checks whether n is a valid complete node.
func (n *Node) validateComplete() error {
	if n.Incomplete() {
		return errors.New("incomplete node")
	}
	if n.UDP == 0 {
		return errors.New("missing UDP port")
	}
	if n.TCP == 0 {
		return errors.New("missing TCP port")
	}
	if n.IP.IsMulticast() || n.IP.IsUnspecified() {
		return errors.New("invalid IP (multicast/unspecified)")
	}
	return nil
}

// String implements fmt.Stringer.
func (n *Node) String() string {
	return string(n.ID) + "@" + n.IP.String() + ":" + strconv.Itoa(int(n.TCP))
}

// Addr returns node's address.
func (n *Node) Addr() string {
	return n.IP.String() + ":" + strconv.Itoa(int(n.TCP))
}

// String returns a string format of NodeID.
func (n NodeID) String() string {
	return string(n)
}

// GenNodeID generates a NodeID.
func GenNodeID() NodeID {
	id := common.ToHex(common.Sha256(common.Int64ToBytes(time.Now().UnixNano())))
	return NodeID(id)
}

// ParseNode parses a string to a Node instance.
func ParseNode(nodeStr string) (node *Node, err error) {
	node = &Node{}
	nodeIDStrs := strings.Split(nodeStr, "@")
	if len(nodeIDStrs) == 2 {
		node.ID = NodeID(nodeIDStrs[0])
		tcpStr := strings.Split(nodeIDStrs[1], ":")
		node.IP = net.ParseIP(tcpStr[0])
		tcp, err := strconv.Atoi(tcpStr[1])
		if err != nil {
			return node, err
		}
		node.TCP = uint16(tcp)
	}
	if len(nodeIDStrs) == 1 {
		tcpStr := strings.Split(nodeIDStrs[0], ":")
		node.IP = net.ParseIP(tcpStr[0])
		tcp, err := strconv.Atoi(tcpStr[1])
		if err != nil {
			return node, err
		}
		node.TCP = uint16(tcp)
	}

	return node, nil

}

// MaxNeighbourNum is the max count of a node's neighbours.
const MaxNeighbourNum = 8
const Threshold = 0.3

// FindNeighbours returns a node's neighbours
func (n *Node) FindNeighbours(ns []*Node) []*Node {
	neighbours := make([]*Node, len(ns))
	for k, v := range ns {
		neighbours[k] = v
	}
	return neighbours
	/*
		if len(ns) < MaxNeighbourNum {
			return ns
		}
		neighbours := make([]*Node, 0)
		witness := params.WitnessNodes
		spnodes := params.SpNodes

		disArr := make([]int, len(ns))
		neighbourKeys := make(map[int]int, 0)
		rand.Seed(time.Now().UnixNano())
		for k, v := range ns {
			if len(neighbours) >= MaxNeighbourNum {
				return neighbours
			}
			if witness[n.Addr()] { // for witness nodes
				if witness[v.Addr()] {
					neighbours = append(neighbours, v)
					neighbourKeys[k] = 1
				}
			} else if spnodes[n.Addr()] { // for sp nodes
				if witness[v.Addr()] {
					neighbours = append(neighbours, v)
					neighbourKeys[k] = 1
				}
			} else { // for ordinary nodes
				if spnodes[v.Addr()] {
					rnd := rand.Float64()
					if rnd > Threshold {
						neighbours = append(neighbours, v)
						neighbourKeys[k] = 1
					}
				}
			}
			disArr[k] = xorDistance(n.Addr(), v.Addr())
		}
		sortArr := make([]int, len(ns))
		copy(sortArr, disArr)
		sort.Ints(sortArr)
		for _, v := range sortArr {
			for k, vd := range disArr {
				if _, ok := neighbourKeys[k]; !ok && v == vd && len(neighbourKeys) < MaxNeighbourNum {
					neighbourKeys[k] = 0
				}
			}
		}
		for k := range neighbourKeys {
			if len(neighbours) >= MaxNeighbourNum || n.Addr() == ns[k].Addr() {
				continue
			}
			if neighbourKeys[k] == 1 {
				continue
			}
			neighbours = append(neighbours, ns[k])
		}
	*/
}

func xorDistance(one, other string) (ret int) {
	oneBytes := []byte(one)
	otherBytes := []byte(other)
	for i := 0; i < len(oneBytes); i++ {
		xor := oneBytes[i] ^ otherBytes[i]
		for j := 0; j < 8; j++ {
			if (xor>>uint8(7-j))&0x01 != 0 {
				return i*8 + j
			}
		}
	}
	return len(oneBytes) * 8
}
