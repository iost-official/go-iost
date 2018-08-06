package p2p

import (
	"context"
	"fmt"
	"net"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
	peer "github.com/libp2p/go-libp2p-peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

type Node struct {
	routeTable *kbucket.RoutingTable
	host       host.Host
}

func basicHost(pk crypto.PrivKey, listen string) (host.Host, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		fmt.Println("failed to resolve tcp addr. err=", err)
		return nil, err
	}
	opts := []libp2p.Option{
		libp2p.Identity(pk),
		libp2p.NATPortMap(),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", tcpAddr.IP, tcpAddr.Port)),
	}
	return libp2p.New(context.Background(), opts...)
}

func NewNode(config *Config) (*Node, error) {
	node := &Node{}

	privKey, err := getOrCreateKey(config.PrivKeyPath)
	if err != nil {
		// node.log.E("failed to get private key. err=%v", err)
		return nil, err
	}

	host, err := basicHost(privKey, config.Listen)
	if err != nil {
		// node.log.E("failed to make a host. err=%v", err)
		return nil, err
	}
	node.host = host

	node.routeTable = kbucket.NewRoutingTable(config.BucketSize, kbucket.ConvertPeerID(node.host.ID()), config.PeerTimeout, node.host.Peerstore())

	return node, nil
}

func (node *Node) Start() error {
	return nil
}

func (node *Node) Stop() error {
	return nil
}

func (node *Node) NeighborAddrs() map[peer.ID][]multiaddr.Multiaddr {
	addrs := make(map[peer.ID][]multiaddr.Multiaddr)
	for _, pid := range node.host.Peerstore().Peers() {
		addrs[pid] = node.host.Peerstore().Addrs(pid)
	}
	return addrs
}
