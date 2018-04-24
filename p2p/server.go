package p2p

import (
	"crypto/ecdsa"
	"net"

	"sync"

	"fmt"

	"time"

	"math/rand"

	"io/ioutil"
	"os"

	"strconv"

	"strings"

	"bufio"
	"bytes"

	"encoding/binary"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core"
	"github.com/iost-official/prototype/iostdb"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/params"
	"github.com/iostio/iost.io/p2p/discover"
	"github.com/pkg/errors"
)

const (
	ReqNodeTableIntervalSecond = 10
	PingIntervalSecond         = 5
	MaxPing                    = 3
	MaxDialRetry               = 3
)

type connFlag int

// Conn represents a connection between two nodes in the network
type conn struct {
	fd net.Conn
	transport
	flags connFlag
	cont  chan error      // The run loop uses cont to signal errors to SetupConn.
	id    discover.NodeID // valid after the encryption handshake
	caps  []Cap           // valid after the protocol handshake
	name  string          // valid after the protocol handshake
}

type transport interface {
	// The two handshakes.
	doEncHandshake(prv *ecdsa.PrivateKey, dialDest *discover.Node) (discover.NodeID, error)
	doProtoHandshake(our *protoHandshake) (*protoHandshake, error)
	// The MsgReadWriter can only be used after the encryption
	// handshake has completed. The code uses conn.id to track this
	// by setting it to a non-nil value after the encryption handshake.
	MsgReadWriter
	// transports must provide Close because we use MsgPipe in some of
	// the tests. Closing the actual network connection doesn't do
	// anything in those tests because NsgPipe doesn't use it.
	close(err error)
}

//接收其他node的数据，定时同步连接远端的路由表，如果ping-pong心跳检测失败，则从本地路由表中随机一个节点，发起连接，
//接收到消息，会广播给其他连接本节点的数据，并通过recv队列通知上层应用
//todo resend
//todo priority broadcast
type Server struct {
	ListenAddr string

	RemoteAddr string
	Conn       net.Conn

	// is seedAddr pinged
	pinged    bool
	seedAddr  string
	nodeTable *iostdb.LDBDatabase //key-address,val-retry times
	lock      sync.RWMutex

	// the nodes which use our server as the remote addr
	peers map[string]net.Conn

	SendCh      chan []byte
	BroadcastCh chan []byte

	RecvCh chan core.Request

	log *log.Logger
}

func NewServer() (*Server, error) {
	send := make(chan []byte, 1)
	recv := make(chan core.Request, 1)
	broadCh := make(chan []byte, 1)
	srvLog, err := log.NewLogger("log_p2p")
	if err != nil {
		fmt.Errorf("failed to init log %v", err)
	}
	dirname, err := ioutil.TempDir(os.TempDir(), "p2p_test_")
	if err != nil {
		fmt.Errorf("failed to init db path %v", err)
	}
	nodeTable, err := iostdb.NewLDBDatabase(dirname, 0, 0)
	if err != nil {
		fmt.Errorf("failed to init db %v", err)
	}
	s := &Server{
		pinged:      false,
		nodeTable:   nodeTable,
		peers:       make(map[string]net.Conn),
		SendCh:      send,
		RecvCh:      recv,
		BroadcastCh: broadCh,
		log:         srvLog,
	}
	return s, nil
}

func (s *Server) Close() {
	defer s.Conn.Close()
}

func (s *Server) Listen(port uint16) (<-chan core.Request, error) {
	s.ListenAddr = "127.0.0.1:" + strconv.Itoa(int(port))
	s.log.D("listening %v", s.ListenAddr)
	l, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return s.RecvCh, errors.New("failed to listen addr, err  = " + fmt.Sprintf("%v", err))
	}
	if isListenAddrNotInBoot(s.ListenAddr) && s.RemoteAddr == "" {
		s.RemoteAddr = s.randBootAddr()
	}
	if s.RemoteAddr != "" {
		s.putNode(s.RemoteAddr)
		s.seedAddr = s.RemoteAddr
	}
	//receive msg
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				s.log.E("accept downStream node err:%v", err)
				continue
			}
			go s.receiveLoop(conn)
		}
	}()
	//send msg
	go s.sendLoop()
	//conn manage
	errConn := s.manageConnLoop()

	return s.RecvCh, errConn
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return errors.New("failed to listen addr, err  = " + fmt.Sprintf("%v", err))
	}
	//receive msg
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				s.log.E("accept downStream node err:%v", err)
				continue
			}
			go s.receiveLoop(conn)
		}
	}()
	//send msg
	go s.sendLoop()
	//conn manage
	return s.manageConnLoop()
}

func (s *Server) Broadcast(r *core.Request) {
	data, err := r.Marshal(nil)
	if err != nil {
		s.log.E("marshal request encountered err:%v", err)
	}
	for downAddr, conn := range s.peers {
		if string(r.From) != downAddr {
			go s.send(conn, data, BroadcastMessage)
		}
	}
}

func (s *Server) Send(r *core.Request) {
	data, err := r.Marshal(nil)
	if err != nil {
		s.log.E("marshal request encountered err:%v", err)
	}
	go s.send(s.Conn, data, Message)
}

func (s *Server) sendLoop() {
	for {
		select {
		case data := <-s.SendCh:
			s.send(s.Conn, data, Message)
		case data := <-s.BroadcastCh:
			req := NewRequest(BroadcastMessage, s.ListenAddr, data)
			s.broadcast(req)
		}
	}
}

func (s *Server) broadcast(r *Request) {
	if r.Type == BroadcastMessage {
		if s.Conn != nil {
			if s.Conn.RemoteAddr().String() != string(r.From) {
				s.send(s.Conn, r.Body, r.Type)
			}
		}

		for downAddr, conn := range s.peers {
			if string(r.From) != downAddr {
				go s.send(conn, r.Body, r.Type)
			}
		}
	}
}
func (s *Server) send(conn net.Conn, body []byte, typ ReqType) {
	if conn == nil {
		s.log.E("from %v,send data = %v, conn is nil", s.ListenAddr, body)
		return
	}

	r := NewRequest(typ, s.ListenAddr, body)
	pack, err := r.Pack()
	if err != nil {
		s.log.E("pack data encountered err:%v", err)
	}
	n, err := conn.Write(pack)
	s.log.D("%v send data: typ= %v, body=%s, n = %v, err : %v", s.ListenAddr, r.Type, string(r.Body), n, err)
}

//接收数据
func (s *Server) receiveLoop(conn net.Conn) {
	var downStreamAddr string
	defer func(addr string) {
		conn.Close()
		if addr != "" {
			if ok, _ := s.nodeTable.Has([]byte(addr)); ok {
				s.nodeTable.Delete([]byte(addr))
			}
		}
	}(downStreamAddr)

	for {
		scanner := bufio.NewScanner(conn)
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if !atEOF && isNetVersionMatch(data) {
				if len(data) > 8 {
					length := int32(0)
					binary.Read(bytes.NewReader(data[4:8]), binary.BigEndian, &length)
					if int(length)+8 <= len(data) {
						return int(length) + 8, data[:int(length)+8], nil
					}
				}
			}
			return
		})
		for scanner.Scan() {
			req := new(Request)
			req.Unpack(bytes.NewReader(scanner.Bytes()))
			addr, err := req.handle(s, conn)
			if err != nil {
				s.log.D("handle message error：", err)
				break
			}
			if addr != "" {
				downStreamAddr = addr
			}
		}
		if err := scanner.Err(); err != nil {
			s.log.E("invalid data packets: %v", err)
			return
		}
	}

	s.log.D("recieve loop finish..")
}

func (s *Server) manageConnLoop() error {
	if s.RemoteAddr != "" {
		s.log.I("start manage conn loop: %v", s)
		go s.ping()
		//request remote node table
		go func() {
			for {
				if s.Conn != nil {
					s.log.D("%v request nodetable", s.ListenAddr)
					s.send(s.Conn, nil, ReqNodeTable)
				}
				time.Sleep(ReqNodeTableIntervalSecond * time.Second)
			}
		}()
		if s.Conn == nil {
			conn, err := net.Dial("tcp", s.RemoteAddr)
			if err != nil {
				s.log.E("dial %v got err: %v", s.RemoteAddr, err)
				return err
			}
			s.Conn = conn
			s.log.I("  conn loop: %+v", s.Conn)
			s.receiveLoop(s.Conn)
			//loop for monitor conn
			go func() {
				dialRetry := 0
				for {
					if s.Conn == nil {
						if dialRetry >= MaxDialRetry {
							if ok, _ := s.nodeTable.Has([]byte(s.seedAddr)); ok {
								s.log.D("delete node: %v", s.seedAddr)
								s.nodeTable.Delete([]byte(s.seedAddr))
							}
							dialRetry = 0
							nodeBytes, _ := s.allNodesExcludeAddr(s.seedAddr)
							addr := s.randNodeAddr(string(nodeBytes))
							if addr != "" {
								s.seedAddr = addr
							}
						}
						conn, err := net.Dial("tcp", s.seedAddr)
						if err != nil {
							s.log.E("monitor: dial %v got err: %v", s, err)
						} else {
							s.Conn = conn
							s.receiveLoop(s.Conn)
						}
						dialRetry++
					}
					time.Sleep(PingIntervalSecond * time.Second)
				}
			}()
		}
	}
	return nil
}

func (s *Server) ping() {
	n := 0
	for {
		if s.pinged {
			n = 0
			s.pinged = false
			continue
		}
		if n >= MaxPing {
			if s.Conn != nil {
				s.Conn.Close()
				s.Conn = nil
				s.seedAddr = ""
			}
			n = 0
			continue
		}

		if s.Conn != nil {
			s.send(s.Conn, []byte(s.ListenAddr), Ping)
			n++
		}
		time.Sleep(PingIntervalSecond * time.Second)
	}
}

func (s *Server) allNodesExcludeAddr(excludeAddr string) ([]byte, error) {
	if s.nodeTable == nil {
		return nil, nil
	}
	addrs := make([]string, 0)
	iter := s.nodeTable.NewIterator()
	for iter.Next() {
		addr := string(iter.Key())
		if addr != excludeAddr {
			addrs = append(addrs, addr)
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, err
	}

	return []byte(strings.Join(addrs, ",")), nil
}

func (s *Server) AllNodes() (string, error) {
	if s.nodeTable == nil {
		return "", nil
	}
	addrs := make([]string, 0)
	iter := s.nodeTable.NewIterator()
	for iter.Next() {
		addr := string(iter.Key())
		addrs = append(addrs, addr)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return "", err
	}
	return strings.Join(addrs, ","), nil
}

func (s *Server) putNode(addr string) {
	if addr != "" && addr != s.ListenAddr {
		s.log.D("%v, append nodetable from remote: %v", s.ListenAddr, addr)
		s.nodeTable.Put([]byte(addr), common.IntToBytes(0))
	}
	return
}

func (s *Server) setPeer(addr string, conn net.Conn) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.peers[addr] = conn
}

func (s *Server) appendNodeTable(addr string) {
	addrs := strings.Split(addr, ",")
	for _, addr := range addrs {
		s.putNode(addr)
	}
}

func (s *Server) randNodeAddr(addr string) string {
	addrs := strings.Split(addr, ",")
	if len(addrs) <= 0 {
		return ""
	}
	rand.Shuffle(rand.Intn(len(addrs)), func(i, j int) { addrs[i], addrs[j] = addrs[j], addrs[i] })
	return addrs[0]
}

func (s *Server) randBootAddr() string {
	addrs := make([]string, 0)
	for _, encodeAddr := range params.TestnetBootnodes {
		addr := extractAddrFromBoot(encodeAddr)
		if addr != "" && s.ListenAddr != addr {
			addrs = append(addrs, addr)
		}
	}
	if len(addrs) <= 0 {
		return ""
	}
	rand.Shuffle(rand.Intn(len(addrs)), func(i, j int) { addrs[i], addrs[j] = addrs[j], addrs[i] })
	return addrs[0]
}

func (s *Server) spreadUp(body []byte) {
	appReq := &core.Request{}
	if _, err := appReq.Unmarshal(body); err == nil {
		s.RecvCh <- *appReq
	}
	return
}

func extractAddrFromBoot(encodeAddr string) string {
	strs := strings.Split(encodeAddr, "@")
	if len(strs) == 2 {
		return strs[1]
	}
	return ""
}

func isListenAddrNotInBoot(listenAddr string) bool {
	for _, addr := range params.TestnetBootnodes {
		if strings.Contains(addr, listenAddr) {
			return false
		}
	}
	return true
}
