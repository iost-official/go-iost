package discover

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"strings"

	"sort"

	"github.com/iost-official/prototype/common"
)

type NodeID string

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

// NewNode creates a new node
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

func (n *Node) String() string {
	return string(n.ID) + "@" + n.IP.String() + ":" + strconv.Itoa(int(n.TCP))
}

func (n *Node) Addr() string {
	return n.IP.String() + ":" + strconv.Itoa(int(n.TCP))
}

// NodeID prints as a long hexadecimal number.
func (n NodeID) String() string {
	return fmt.Sprintf("%s", string(n))
}

func GenNodeId() NodeID {
	id := common.ToHex(common.Sha256(common.Int64ToBytes(time.Now().UnixNano())))
	return NodeID(id)
}

func ParseNode(nodeStr string) (node *Node, err error) {
	node = &Node{}
	nodeIdStrs := strings.Split(nodeStr, "@")
	if len(nodeIdStrs) == 2 {
		node.ID = NodeID(nodeIdStrs[0])
		tcpStr := strings.Split(nodeIdStrs[1], ":")
		node.IP = net.ParseIP(tcpStr[0])
		tcp, err := strconv.Atoi(tcpStr[1])
		if err != nil {
			return node, err
		}
		node.TCP = uint16(tcp)
	}
	if len(nodeIdStrs) == 1 {
		tcpStr := strings.Split(nodeIdStrs[0], ":")
		node.IP = net.ParseIP(tcpStr[0])
		tcp, err := strconv.Atoi(tcpStr[1])
		if err != nil {
			return node, err
		}
		node.TCP = uint16(tcp)
	}

	return node, nil

}

const MaxNeighbourNum = 8

func (n *Node) FindNeighbours(ns []*Node) []*Node {
	if len(ns) < MaxNeighbourNum {
		return ns
	}
	neighbours := make([]*Node, 0)
	disArr := make([]int, len(ns))
	for k, v := range ns {
		disArr[k] = xorDistance(n.ID, v.ID)
	}
	sortArr := make([]int, len(ns))
	copy(sortArr, disArr)
	sort.Ints(sortArr)

	neighbourKeys := make(map[int]int, 0)
	for _, v := range sortArr {
		for k, vd := range disArr {
			if _, ok := neighbourKeys[k]; !ok && v == vd && len(neighbourKeys) < MaxNeighbourNum {
				neighbourKeys[k] = 0
			}
		}
	}
	for k, _ := range neighbourKeys {
		if len(neighbours) >= MaxNeighbourNum || n.Addr() == ns[k].Addr() {
			break
		}
		neighbours = append(neighbours, ns[k])
	}
	return neighbours
}

func xorDistance(one, other NodeID) (ret int) {
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
