package p2p

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/iost-official/Go-IOS-Protocol/ilog"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
)

// PeerID is the alias of peer.ID
type PeerID = peer.ID

const (
	protocolID = "iostp2p/1.0"
)

// errors
var (
	ErrPortUnavailable = errors.New("port is unavailable")
)

// Service defines all the API of p2p package.
type Service interface {
	Start() error
	Stop()

	Broadcast([]byte, MessageType, MessagePriority)
	SendToPeer(PeerID, []byte, MessageType, MessagePriority)
	Register(string, ...MessageType) chan IncomingMessage
	Deregister(string, ...MessageType)
}

// NetService is the implementation of Service interface.
type NetService struct {
	host        host.Host
	peerManager *PeerManager
	config      *Config
}

// NewDefault returns a default NetService instance.
func NewDefault() (*NetService, error) {
	ns := &NetService{}
	privKey, err := getOrCreateKey("priv.key")
	if err != nil {
		ilog.Error("failed to get private key. err=%v", err)
		return nil, err
	}
	host, err := ns.startHost(privKey, "0.0.0.0:6666")
	if err != nil {
		ilog.Error("failed to start a host. err=%v", err)
		return nil, err
	}
	ns.host = host

	ns.peerManager = NewPeerManager(host, DefaultConfig())

	return ns, nil
}

// NewNetService returns a NetService instance with the config argument.
func NewNetService(config *Config) (*NetService, error) {
	ns := &NetService{
		config: config,
	}

	privKey, err := getOrCreateKey(config.PrivKeyPath)
	if err != nil {
		ilog.Error("failed to get private key. err=%v, path=%v", err, config.PrivKeyPath)
		return nil, err
	}

	host, err := ns.startHost(privKey, config.ListenAddr)
	if err != nil {
		ilog.Error("failed to start a host. err=%v, listenAddr=%v", err, config.ListenAddr)
		return nil, err
	}
	ns.host = host

	ns.peerManager = NewPeerManager(host, config)

	return ns, nil
}

// Start starts the jobs.
func (ns *NetService) Start() error {
	ns.peerManager.Start()
	return nil
}

// Stop stops all the jobs.
func (ns *NetService) Stop() {
	ns.host.Close()
	ns.peerManager.Stop()
	return
}

// Broadcast broadcasts the data.
func (ns *NetService) Broadcast(data []byte, typ MessageType, mp MessagePriority) {
	ns.peerManager.Broadcast(data, typ, mp)
}

// SendToPeer sends data to the given peer.
func (ns *NetService) SendToPeer(peerID peer.ID, data []byte, typ MessageType, mp MessagePriority) {
	ns.peerManager.SendToPeer(peerID, data, typ, mp)
}

// Register registers a message channel of the given types.
func (ns *NetService) Register(id string, typs ...MessageType) chan IncomingMessage {
	return ns.peerManager.Register(id, typs...)
}

// Deregister deregisters a message channel of the given types.
func (ns *NetService) Deregister(id string, typs ...MessageType) {
	ns.peerManager.Deregister(id, typs...)
}

// startHost starts a libp2p host.
func (ns *NetService) startHost(pk crypto.PrivKey, listenAddr string) (host.Host, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		fmt.Println("failed to resolve tcp addr. err=", err)
		return nil, err
	}

	if !isPortAvailable(tcpAddr.Port) {
		return nil, ErrPortUnavailable
	}

	opts := []libp2p.Option{
		libp2p.Identity(pk),
		libp2p.NATPortMap(),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", tcpAddr.IP, tcpAddr.Port)),
	}
	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		ilog.Error("failed to start host. err=%v", err)
		return nil, err
	}
	h.SetStreamHandler(protocolID, ns.streamHandler)
	return h, nil
}

func (ns *NetService) streamHandler(s libnet.Stream) {
	ns.peerManager.HandleStream(s)
}
