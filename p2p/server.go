package p2p

import (
	"net"

	"errors"
	"sync"

	"fmt"

	"strconv"

	"github.com/iost-official/prototype/iostdb"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/p2p/discover"
	"github.com/iostio/iost.io/p2p/nat"
)

type Config struct {
	// MaxPeers is the maximum number of peers that can be
	// connected. It must be greater than zero.
	MaxPeers int

	// Name sets the node name of this server.
	// Use common.MakeName to create a name that follows existing conventions.
	Name       string
	ListenAddr string
	NAT        nat.Interface `toml:",omitempty"`
	// NodeDatabase is the path to the database containing the previously seen
	// live nodes in the network.
	NodeDatabase string
}

// Server manages all peer connections.
type Server struct {
	// Config fields may not be modified while the server is running.
	Config
	nodeTable *iostdb.LDBDatabase //database of known nodes
	lock      sync.RWMutex        // protects running
	running   bool

	listener net.Listener

	quit       chan struct{}
	addPeer    chan *Peer
	delPeer    chan *Peer
	dialingMap map[string]bool
	loopWG     sync.WaitGroup // loop, listenLoop
	log        *log.Logger

	PeerSet *peerSet //like peers pool
}

// Start starts running the server.
func (srv *Server) Start() (err error) {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if srv.running {
		return errors.New("server already running")
	}
	srv.running = true
	if srv.log == nil {
		srv.log, _ = log.NewLogger("p2p ")
	}
	srv.log.I("Starting P2P networking")

	srv.quit = make(chan struct{})
	srv.addPeer = make(chan *Peer)
	srv.delPeer = make(chan *Peer)
	srv.PeerSet = &peerSet{}

	srv.loopWG.Add(1)
	if srv.ListenAddr != "" {
		if err := srv.startListening(); err != nil {
			return err
		}
	}
	go srv.run()
	srv.running = true
	return nil
}

func (srv *Server) startListening() error {
	// Launch the TCP listener.
	listener, err := net.Listen("tcp", srv.ListenAddr)
	if err != nil {
		return err
	}
	laddr := listener.Addr().(*net.TCPAddr)
	srv.ListenAddr = laddr.String()
	srv.listener = listener
	srv.loopWG.Add(1)
	go srv.listenLoop()
	// Map the TCP listening port if NAT is configured.
	if !laddr.IP.IsLoopback() && srv.NAT != nil {
		srv.loopWG.Add(1)
		go func() {
			nat.Map(srv.NAT, srv.quit, "tcp", laddr.Port, laddr.Port, "iost p2p")
			srv.loopWG.Done()
		}()
	}
	return nil
}

func (srv *Server) listenLoop() {
	defer srv.loopWG.Done()
	for {
		var (
			fd  net.Conn
			err error
		)
		for {
			fd, err = srv.listener.Accept()
			if tempErr, ok := err.(tempError); ok && tempErr.Temporary() {
				srv.log.D("Temporary read error", "err", err)
				continue
			} else if err != nil {
				srv.log.D("Read error", "err", err)
				return
			}
			break
		}

		go func() {
			n := &NetworkImpl{
				db:   srv.nodeTable,
				done: false,
				conn: fd,
			}

			node, err := discover.Addr2Node(n.conn.RemoteAddr().String())
			if err != nil {
				srv.log.E("addr2Node got err", err)
			}
			srv.addPeer <- newPeer(n, node)
		}()
	}
}

func (srv *Server) run() {
	defer srv.loopWG.Done()
running:
	for {
		if len(srv.PeerSet.peers)+len(srv.dialingMap) < srv.MaxPeers {
			srv.scheduleTask()
		}
		select {
		case <-srv.quit:
			srv.log.D("server closed")
			break running
		case p := <-srv.addPeer:
			srv.log.D("Adding peer", fmt.Sprintf("%v", srv), fmt.Sprintf(" ::: %+v", p.rw.conn))
			srv.PeerSet.Set(p.to.String(), p)
		case p := <-srv.delPeer:
			srv.log.D("deleting peer", p)
			if p := srv.PeerSet.Get(p.to.ID.String()); p != nil {
				p.rw.conn.Close()
			}
			srv.PeerSet.Remove(p.to.ID.String())
		}
	}
	srv.log.D("close all peers in p2p")
	for _, p := range srv.PeerSet.peers {
		p.rw.conn.Close()
	}
}

//create net conn
func (srv *Server) scheduleTask() {
	nodesNotInPeers := make([]*discover.Node, 0)
	nodes := srv.AllNodes()
	for _, n := range nodes {
		if p := srv.PeerSet.Get(n.ID.String()); p == nil {
			if _, dOk := srv.dialingMap[n.ID.String()]; !dOk {
				nodesNotInPeers = append(nodesNotInPeers, n)
			}
		}
	}
	node := discover.GetRandomNode(nodesNotInPeers)
	srv.log.D("fetch rand node " + fmt.Sprintf("%v", node))
	if node == nil {
		srv.log.E("node table is empty")
		return
	}

	go func(node *discover.Node, srv *Server) {
		n := &NetworkImpl{
			db:   srv.nodeTable,
			done: false,
		}
		var err error
		n.conn, err = net.Dial("tcp", string(node.IP)+":"+strconv.Itoa(int(node.TCP)))
		if err != nil {
			srv.log.E("dial got err: " + fmt.Sprintf("%+v", err))
		}

		srv.log.D("net dial " + fmt.Sprintf("%+v", n))
		srv.addPeer <- newPeer(n, node)
	}(node, srv)
}

func (srv *Server) AllNodes() []*discover.Node {
	srv.lock.RLock()
	defer srv.lock.RUnlock()
	nodes := make([]*discover.Node, 0)

	iter := srv.nodeTable.NewIterator()
	for iter.Next() {
		addr, _ := srv.nodeTable.Get([]byte(string(iter.Key())))
		node, _ := discover.Addr2Node(string(addr))
		if node != nil {
			nodes = append(nodes, node)
		}

	}
	iter.Release()
	if err := iter.Error(); err != nil {
		srv.log.E("nodes iterator got err: %v", err)
	}
	return nodes
}

func (srv *Server) Stop() {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if !srv.running {
		return
	}
	srv.running = false
	if srv.listener != nil {
		srv.listener.Close()
	}
	close(srv.quit)
	srv.loopWG.Wait()
}

type tempError interface {
	Temporary() bool
}
