package p2p

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/ilog"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	libnet "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"

	multiaddr "github.com/multiformats/go-multiaddr"
)

//go:generate mockgen --build_flags=--mod=mod -destination mocks/mock_service.go -package p2p_mock github.com/iost-official/go-iost/v3/p2p Service

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

	ID() string
	ConnectBPs([]string)
	PutPeerToBlack(string)

	Broadcast([]byte, MessageType, MessagePriority)
	SendToPeer(PeerID, []byte, MessageType, MessagePriority)
	Register(string, ...MessageType) chan IncomingMessage
	Deregister(string, ...MessageType)

	GetAllNeighbors() []*Peer
}

// NetService is the implementation of Service interface.
type NetService struct {
	*PeerManager

	host        host.Host
	adminServer *adminServer
	config      *common.P2PConfig
}

var _ Service = &NetService{}

// NewNetService returns a NetService instance with the config argument.
func NewNetService(config *common.P2PConfig) (*NetService, error) {
	ns := &NetService{
		config: config,
	}

	if err := os.MkdirAll(config.DataPath, 0755); config.DataPath != "" && err != nil {
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

	ns.PeerManager = NewPeerManager(host, config)

	ns.adminServer = newAdminServer(config.AdminPort, ns.PeerManager)

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
	go ns.PeerManager.Start()
	go ns.adminServer.Start()
	for _, addr := range ns.LocalAddrs() {
		ilog.Infof("local multiaddr: %s/ipfs/%s", addr, ns.ID())
	}
	return nil
}

// Stop stops all the jobs.
func (ns *NetService) Stop() {
	ns.host.Close()
	ns.adminServer.Stop()
	ns.PeerManager.Stop()
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

	secOptions := []libp2p.Option{
		libp2p.Security(tls.ID, tls.New),
		libp2p.Security(noise.ID, noise.New),
	}
	muxOptions := []libp2p.Option{
		libp2p.Muxer(yamux.ID, yamux.DefaultTransport),
	}
	if !ns.config.DisableMplex {
		muxOptions = append(muxOptions, libp2p.Muxer(protocolID, mplex.DefaultTransport))
	}
	opts := []libp2p.Option{
		libp2p.Identity(pk),
		libp2p.NATPortMap(),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", tcpAddr.IP, tcpAddr.Port)),
		libp2p.ChainOptions(muxOptions...),
		libp2p.ChainOptions(secOptions...),
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}
	h.SetStreamHandler(protocolID, ns.streamHandler)
	h.SetStreamHandler(yamux.ID, ns.streamHandler)
	return h, nil
}

func (ns *NetService) streamHandler(s libnet.Stream) {
	ns.PeerManager.HandleStream(s, inbound)
}
