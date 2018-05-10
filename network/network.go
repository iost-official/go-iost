package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	"sync"

	"bufio"

	"strings"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/network/discover"
	"github.com/iost-official/prototype/params"
)

type RequestHead struct {
	Length uint32 // length of Request
}

const (
	HEADLENGTH               = 4
	CheckKnownNodeInterval   = 10
	NodeLiveThresholdSeconds = 30
)

type Response struct {
	From        string
	To          string
	Code        int    // like http status code
	Description string // code status description
}

//Network api
type Network interface {
	Broadcast(req message.Message)
	Send(req message.Message)
	Listen(port uint16) (<-chan message.Message, error)
	Close(port uint16) error
}

type NaiveNetwork struct {
	db     *db.LDBDatabase //database of known nodes
	listen net.Listener
	conn   net.Conn
	done   bool
}

//NewNaiveNetwork create n peers
func NewNaiveNetwork(n int) (*NaiveNetwork, error) {
	dirname, err := ioutil.TempDir(os.TempDir(), "p2p_test_")
	if err != nil {
		return nil, err
	}
	db, err := db.NewLDBDatabase(dirname, 0, 0)
	if err != nil {
		return nil, err
	}
	nn := &NaiveNetwork{
		db:     db,
		listen: nil,
		done:   false,
	}
	for i := 1; i <= n; i++ {
		nn.db.Put([]byte(string(i)), []byte("127.0.0.1:"+strconv.Itoa(11036+i)))
	}
	return nn, nil
}

func (nn *NaiveNetwork) Close(port uint16) error {
	port = 3 // 避免出现unused variable
	nn.done = true
	return nn.listen.Close()
}

//
func (nn *NaiveNetwork) Broadcast(req message.Message) error {
	iter := nn.db.NewIterator()
	for iter.Next() {
		addr, _ := nn.db.Get([]byte(string(iter.Key())))
		conn, err := net.Dial("tcp", string(addr))
		if err != nil {
			if dErr := nn.db.Delete([]byte(string(iter.Key()))); dErr != nil {
				fmt.Errorf("failed to delete peer : k= %v, v = %v, err:%v\n", iter.Key(), addr, err.Error())
			}
			fmt.Errorf("dialing to %v encounter err : %v\n", addr, err.Error())
			continue
		}
		nn.conn = conn
		go nn.Send(req)
	}
	iter.Release()
	return iter.Error()
}

func (nn *NaiveNetwork) Send(req message.Message) {
	if nn.conn == nil {
		fmt.Errorf("no connect in network")
		return
	}
	defer nn.conn.Close()

	reqHeadBytes, reqBodyBytes, err := reqToBytes(req)
	if err != nil {
		fmt.Errorf("reqToBytes encounter err : %v\n", err.Error())
		return
	}
	if _, err = nn.conn.Write(reqHeadBytes); err != nil {
		fmt.Errorf("sending request head encounter err :%v\n", err.Error())
	}
	if _, err = nn.conn.Write(reqBodyBytes[:]); err != nil {
		fmt.Errorf("sending request body encounter err : %v\n", err.Error())
	}
	return
}

func (nn *NaiveNetwork) Listen(port uint16) (<-chan message.Message, error) {
	var err error
	nn.listen, err = net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		return nil, fmt.Errorf("Error listening: %v", err.Error())
	}
	fmt.Println("Listening on " + ":" + strconv.Itoa(int(port)))
	req := make(chan message.Message)

	conn := make(chan net.Conn)

	// For every listener spawn the following routine
	go func(l net.Listener) {
		for {
			c, err := l.Accept()
			if err != nil {
				// handle error
				conn <- nil
				return
			}
			conn <- c
		}
	}(nn.listen)
	go func() {
		for {
			select {
			case c := <-conn:
				if c == nil {
					if nn.done {
						return
					}
					fmt.Println("Error accepting: ")
					break
				}

				go func(conn net.Conn) {
					defer conn.Close()
					// Make a buffer to hold incoming data.
					buf := make([]byte, HEADLENGTH)
					// Read the incoming connection into the buffer.
					_, err := conn.Read(buf)
					if err != nil {
						fmt.Errorf("Error reading request head:%v", err.Error())
					}
					length := binary.BigEndian.Uint32(buf)
					_buf := make([]byte, length)
					_, err = conn.Read(_buf)

					if err != nil {
						fmt.Errorf("Error reading request body:%v", err.Error())
					}
					var received message.Message
					received.Unmarshal(_buf)
					req <- received
					// Send a response back to person contacting us.
					//conn.Write([]byte("Message received."))
				}(c)
			case <-time.After(1000.0 * time.Second):
				fmt.Println("accepting time out..")
			}
		}
	}()
	return req, nil
}

func reqToBytes(req message.Message) ([]byte, []byte, error) {
	reqBodyBytes, err := req.Marshal(nil)
	if err != nil {
		return nil, reqBodyBytes, err
	}
	reqHead := new(bytes.Buffer)
	if err := binary.Write(reqHead, binary.BigEndian, int32(len(reqBodyBytes))); err != nil {
		return nil, reqBodyBytes, err
	}
	return reqHead.Bytes(), reqBodyBytes, nil
}

type NetConifg struct {
	LogPath       string
	NodeTablePath string
}

func (conf *NetConifg) SetLogPath(path string) *NetConifg {
	if path == "" {
		fmt.Errorf("path of log should not be empty")
	}
	conf.LogPath = path
	return conf
}

func (conf *NetConifg) SetNodeTablePath(path string) *NetConifg {
	if path == "" {
		fmt.Errorf("path of node table should not be empty")
	}
	conf.NodeTablePath = path
	return conf
}

//BaseNetwork boot node maintain all node table, and distribute the node table to all node
type BaseNetwork struct {
	nodeTable  *db.LDBDatabase //all known node except remoteAddr
	neighbours map[string]*discover.Node
	lock       sync.RWMutex

	RecvCh chan message.Message

	recentSentMap map[string]message.Message
	localNode     *discover.Node
	log           *log.Logger
}

// NewBaseNetwork ...
func NewBaseNetwork(conf *NetConifg) (*BaseNetwork, error) {
	recv := make(chan message.Message, 1)
	if conf.LogPath == "" {
		conf.LogPath, _ = ioutil.TempDir(os.TempDir(), "iost_log_")
	}
	if conf.NodeTablePath == "" {
		conf.NodeTablePath, _ = ioutil.TempDir(os.TempDir(), "iost_node_table_")
	}
	srvLog, err := log.NewLogger(conf.LogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init log %v", err)
	}
	_, pErr := os.Stat(conf.NodeTablePath)
	if pErr != nil {
		return nil, fmt.Errorf("failed to init db path %v", pErr)

	}
	nodeTable, err := db.NewLDBDatabase(conf.NodeTablePath, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to init db %v", err)
	}
	sentMap := make(map[string]message.Message, 0)
	localNode := &discover.Node{ID: discover.GenNodeId(), IP: []byte("127.0.0.1")}
	neighbours := make(map[string]*discover.Node, 0)
	s := &BaseNetwork{
		nodeTable:     nodeTable,
		RecvCh:        recv,
		localNode:     localNode,
		recentSentMap: sentMap,
		neighbours:    neighbours,
		log:           srvLog,
	}
	return s, nil
}

// Listen listen local port, find neighbours
func (bn *BaseNetwork) Listen(port uint16) (<-chan message.Message, error) {
	bn.localNode.TCP = port
	bn.log.D("listening %v", bn.localNode.Addr())
	l, err := net.Listen("tcp", bn.localNode.Addr())
	if err != nil {
		return bn.RecvCh, errors.New("failed to listen addr, err  = " + fmt.Sprintf("%v", err))
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				bn.log.E("accept downStream node err:%v", err)
				continue
			}
			go bn.receiveLoop(conn)
		}
	}()
	//register
	go bn.registerLoop()
	go bn.nodeCheckLoop()
	go bn.cleanRecentSentLoop()
	return bn.RecvCh, nil
}

//Broadcast msg to all node in the node table
func (bn *BaseNetwork) Broadcast(msg message.Message) {
	neighbours := bn.neighbours
	for _, node := range neighbours {
		bn.log.D("broad msg: %v to node: %v", msg, node.String())
		msg.To = node.String()
		go bn.broadcast(msg)
	}
}

//broadcast broadcast to all neighbours, stop broadcast when msg already broadcast
func (bn *BaseNetwork) broadcast(msg message.Message) {
	data, err := msg.Marshal(nil)
	if err != nil {
		bn.log.E("marshal request encountered err:%v", err)
	}
	msgHash := common.Base58Encode(common.Sha256(msg.Body))
	if _, ok := bn.recentSentMap[msgHash]; !ok {
		bn.recentSentMap[msgHash] = msg
	} else {
		return
	}
	node, _ := discover.ParseNode(msg.To)
	conn, err := net.Dial("tcp", node.Addr())
	if err != nil {
		bn.log.E("dial tcp %v got err:%v", node.Addr(), err)
	}
	req := newRequest(BroadcastMessage, bn.localNode.String(), data)
	defer conn.Close()
	bn.send(conn, req)
}

//Send msg to msg.To
func (bn *BaseNetwork) Send(msg message.Message) {
	conn, err := net.Dial("tcp", msg.To)
	if err != nil {
		bn.log.E("dial tcp %v got err:%v", msg.To, err)
	}
	data, err := msg.Marshal(nil)
	if err != nil {
		bn.log.E("marshal request encountered err:%v", err)
	}
	req := newRequest(Message, bn.localNode.String(), data)
	defer conn.Close()
	bn.send(conn, req)
}

// Close all connection
func (bn *BaseNetwork) Close(port uint16) error {
	return nil
}

func (bn *BaseNetwork) send(conn net.Conn, r *Request) {
	if conn == nil {
		bn.log.E("from %v,send data = %v, conn is nil", bn.localNode.String(), r)
		return
	}
	pack, err := r.Pack()
	if err != nil {
		bn.log.E("pack data encountered err:%v", err)
	}
	n, err := conn.Write(pack)
	bn.log.D("%v send data: typ= %v, body=%s, n = %v, err : %v", bn.localNode.String(), r.Type, string(r.Body), n, err)
}

func (bn *BaseNetwork) receiveLoop(conn net.Conn) {
	defer conn.Close()
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
			req.response(bn, conn)
		}
		if err := scanner.Err(); err != nil {
			bn.log.E("invalid data packets: %v", err)
			return
		}
	}
	bn.log.D("recieve loop finish..")
}

//AllNodesExcludeAddr returns all the known node in the network
func (bn *BaseNetwork) AllNodesExcludeAddr(excludeAddr string) ([]string, error) {
	if bn.nodeTable == nil {
		return nil, nil
	}
	addrs := make([]string, 0)
	iter := bn.nodeTable.NewIterator()
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

	return addrs, nil
}

//put node into node table of server
func (bn *BaseNetwork) putNode(addrs string) {
	addrArr := strings.Split(addrs, ",")
	for _, addr := range addrArr {
		if addr != "" && addr != bn.localNode.String() {
			bn.nodeTable.Put([]byte(addr), common.Int64ToBytes(time.Now().Unix()))
		}
	}
	bn.findNeighbours()
	return
}

//nodeCheckLoop inspection Last registration time of node
func (bn *BaseNetwork) nodeCheckLoop() {
	for {
		now := time.Now().Unix()
		iter := bn.nodeTable.NewIterator()
		for iter.Next() {
			if (now - common.BytesToInt64(iter.Value())) > NodeLiveThresholdSeconds {
				bn.log.D("delete node %v, cuz its last register time is %v", common.BytesToInt64(iter.Value()))
				bn.nodeTable.Delete(iter.Key())
			}
		}
		time.Sleep(CheckKnownNodeInterval * time.Second)
	}
}

//registerLoop register local address to boot nodes
func (bn *BaseNetwork) registerLoop() {
	for {
		for _, encodeAddr := range params.TestnetBootnodes {
			addr := extractAddrFromBoot(encodeAddr)
			if addr != "" && bn.localNode.String() != addr {
				req := newRequest(ReqNodeTable, bn.localNode.String(), nil)
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					bn.log.E("failed to connect boot node, err:%v", err)
				}
				defer conn.Close()
				go bn.receiveLoop(conn)
				bn.log.D("%v request node table from %v", bn.localNode.String(), addr)
				bn.send(conn, req)
			}
		}
		time.Sleep(CheckKnownNodeInterval * time.Second)
	}
}

const validitySentSeconds = 90

//cleanRecentSentLoop
func (bn *BaseNetwork) cleanRecentSentLoop() {
	for {
		msgs := bn.recentSentMap
		now := time.Now().UnixNano()
		for k, msg := range msgs {
			if (now-msg.Time)/1e9 > validitySentSeconds {
				bn.lock.Lock()
				delete(bn.recentSentMap, k)
				bn.lock.Unlock()
			}
		}
		time.Sleep(validitySentSeconds * time.Second)
	}
}
func (bn *BaseNetwork) findNeighbours() {
	nodesStr, _ := bn.AllNodesExcludeAddr(bn.localNode.String())
	nodes := make([]*discover.Node, 0)
	for _, nodeStr := range nodesStr {
		node, _ := discover.ParseNode(nodeStr)
		nodes = append(nodes, node)
	}
	neighbours := bn.localNode.FindNeighbours(nodes)
	for _, n := range neighbours {
		bn.setNeighbour(n)
	}
}

func (bn *BaseNetwork) setNeighbour(node *discover.Node) {
	bn.lock.Lock()
	defer bn.lock.Unlock()
	bn.neighbours[node.String()] = node
}
