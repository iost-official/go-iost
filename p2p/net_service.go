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

type Service interface {
	Start() error
	Stop()

	Broadcast([]byte, MessageType, MessagePriority)
	SendToPeer(peer.ID, []byte, MessageType, MessagePriority)
	Register(string, ...MessageType) chan IncomingMessage
	Deregister(string, ...MessageType)
}

type NetService struct {
	routeTable  *kbucket.RoutingTable
	host        host.Host
	peerManager *PeerManager
}

func basicHost(pk crypto.PrivKey, listenAddr string) (host.Host, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
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

func NewNetService(config *Config) (*NetService, error) {
	ns := &NetService{}

	privKey, err := getOrCreateKey(config.PrivKeyPath)
	if err != nil {
		// node.log.E("failed to get private key. err=%v", err)
		return nil, err
	}

	host, err := basicHost(privKey, config.ListenAddr)
	if err != nil {
		// node.log.E("failed to make a host. err=%v", err)
		return nil, err
	}
	ns.host = host

	ns.routeTable = kbucket.NewRoutingTable(config.BucketSize, kbucket.ConvertPeerID(ns.host.ID()), config.PeerTimeout, ns.host.Peerstore())

	return ns, nil
}

func (ns *NetService) Start() error {
	return nil
}

func (ns *NetService) Stop() error {
	return nil
}

func (ns *NetService) Broadcast(data []byte, typ MessageType, mp MessagePriority) {
	ns.peerManager.Broadcast(data, typ, mp)
}

func (ns *NetService) SendToPeer(peerID peer.ID, data []byte, typ MessageType, mp MessagePriority) {
	ns.peerManager.SendToPeer(peerID, data, typ, mp)
}

func (ns *NetService) Register(id string, typs ...MessageType) chan IncomingMessage {
	return ns.peerManager.Register(id, typs...)
}

func (ns *NetService) Deregister(id string, typs ...MessageType) {
	ns.peerManager.Deregister(id, typs...)
}

func (ns *NetService) NeighborAddrs() map[peer.ID][]multiaddr.Multiaddr {
	addrs := make(map[peer.ID][]multiaddr.Multiaddr)
	for _, pid := range ns.host.Peerstore().Peers() {
		addrs[pid] = ns.host.Peerstore().Addrs(pid)
	}
	return addrs
}
