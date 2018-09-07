package p2p

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	multiaddr "github.com/multiformats/go-multiaddr"
)

//go:generate mockgen -destination mocks/mock_service.go -package p2p_mock github.com/iost-official/Go-IOS-Protocol/p2p Service

// PeerID is the alias of peer.ID
type PeerID = peer.ID

const (
	protocolID  = "iostp2p/1.0"
	privKeyFile = "priv.key"
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
	config      *common.P2PConfig
}

var _ Service = &NetService{}

// NewNetService returns a NetService instance with the config argument.
func NewNetService(config *common.P2PConfig) (*NetService, error) {
	ns := &NetService{
		config: config,
	}

	if err := os.MkdirAll(config.DataPath, 0766); err != nil {
		ilog.Errorf("failed to create p2p datapath, err=%v, path=%v", err, config.DataPath)
		return nil, err
	}

	privKey, err := getOrCreateKey(filepath.Join(config.DataPath, privKeyFile))
	if err != nil {
		ilog.Errorf("failed to get private key. err=%v, path=%v", err, config.DataPath)
		return nil, err
	}

	host, err := ns.startHost(privKey, config.ListenAddr)
	if err != nil {
		ilog.Errorf("failed to start a host. err=%v, listenAddr=%v", err, config.ListenAddr)
		return nil, err
	}
	ns.host = host

	ns.peerManager = NewPeerManager(host, config)

	return ns, nil
}

// ID returns the host's ID.
func (ns *NetService) ID() string {
	return ns.host.ID().Pretty()
}

// LocalAddrs returns the local's multiaddrs.
func (ns *NetService) LocalAddrs() []multiaddr.Multiaddr {
	return ns.host.Addrs()
}

// Start starts the jobs.
func (ns *NetService) Start() error {
	go ns.peerManager.Start()
	for _, addr := range ns.host.Addrs() {
		ilog.Infof("multiaddr: %s/ipfs/%s", addr, ns.ID())
	}
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
		return nil, err
	}
	h.SetStreamHandler(protocolID, ns.streamHandler)
	return h, nil
}

func (ns *NetService) streamHandler(s libnet.Stream) {
	ns.peerManager.HandleStream(s)
}
