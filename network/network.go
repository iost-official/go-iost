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

	"math/rand"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/db"
	"github.com/iost-official/prototype/log"
	"github.com/iost-official/prototype/network/discover"
)

const (
	HEADLENGTH               = 4
	CheckKnownNodeInterval   = 10
	NodeLiveThresholdSeconds = 20
	MaxDownloadRetry         = 2
)

//Network api
type Network interface {
	Broadcast(req message.Message)
	Send(req message.Message)
	Listen(port uint16) (<-chan message.Message, error)
	Close(port uint16) error
	Download(start, end uint64) error
	CancelDownload(start, end uint64) error
}

//NetConfig p2p net config
type NetConifg struct {
	LogPath       string
	NodeTablePath string
	NodeID        string
	ListenAddr    string
	RegisterAddr  string
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

func (conf *NetConifg) SetNodeID(id string) *NetConifg {
	if id == "" {
		fmt.Errorf("node id should not be empty")
	}
	conf.NodeID = id
	return conf
}

func (conf *NetConifg) SetListenAddr(addr string) *NetConifg {
	if addr == "" {
		fmt.Errorf("listen addr should not be empty")
	}
	conf.ListenAddr = addr
	return conf
}

//BaseNetwork boot node maintain all node table, and distribute the node table to all node
type BaseNetwork struct {
	nodeTable  *db.LDBDatabase //all known node except remoteAddr
	neighbours map[string]*discover.Node
	lock       sync.Mutex
	peers      peerSet // manage all connection
	RecvCh     chan message.Message
	listener   net.Listener

	NodeHeightMap map[string]uint64 //maintain all height of nodes higher than current height
	localNode     *discover.Node

	DownloadHeights map[uint64]uint8 //map[height]retry_times
	regAddr         string
	log             *log.Logger
}

// NewBaseNetwork ...
func NewBaseNetwork(conf *NetConifg) (*BaseNetwork, error) {
	recv := make(chan message.Message, 100)
	var err error
	if conf.LogPath == "" {
		conf.LogPath, err = ioutil.TempDir(os.TempDir(), "iost_log_")
		if err != nil {
			return nil, fmt.Errorf("iost_log_path err: %v", err)
		}
	}
	if conf.NodeTablePath == "" {
		conf.NodeTablePath = "iost_node_table_"
	}
	srvLog, err := log.NewLogger(conf.LogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init log %v", err)
	}
	nodeTable, err := db.NewLDBDatabase(conf.NodeTablePath, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to init db %v", err)
	}
	neighbours := make(map[string]*discover.Node, 0)
	NodeHeightMap := make(map[string]uint64, 0)
	if conf.NodeID == "" {
		conf.NodeID = string(discover.GenNodeId())
	}
	localNode := &discover.Node{ID: discover.NodeID(conf.NodeID), IP: net.ParseIP(conf.ListenAddr)}
	downloadHeights := make(map[uint64]uint8, 0)
	s := &BaseNetwork{
		nodeTable:       nodeTable,
		RecvCh:          recv,
		localNode:       localNode,
		neighbours:      neighbours,
		log:             srvLog,
		NodeHeightMap:   NodeHeightMap,
		DownloadHeights: downloadHeights,
		regAddr:         conf.RegisterAddr,
	}
	return s, nil
}

// Listen listen local port, find neighbours
func (bn *BaseNetwork) Listen(port uint16) (<-chan message.Message, error) {
	bn.localNode.TCP = port
	bn.log.D("[net] listening %v", bn.localNode)
	var err error
	bn.listener, err = net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(int(bn.localNode.TCP)))
	if err != nil {
		return bn.RecvCh, errors.New("failed to listen addr, err  = " + fmt.Sprintf("%v", err))
	}
	go func() {
		for {
			conn, err := bn.listener.Accept()
			if err != nil {
				bn.log.E("[net] accept downStream node err:%v", err)
				continue
			}
			go bn.receiveLoop(conn)
		}
	}()
	//register
	go bn.registerLoop()
	go bn.nodeCheckLoop()
	return bn.RecvCh, nil
}

//Broadcast msg to all node in the node table
func (bn *BaseNetwork) Broadcast(msg message.Message) {
	neighbours := bn.neighbours
	if msg.From == "" {
		msg.From = bn.localNode.Addr()
	}
	from := msg.From
	for _, node := range neighbours {
		bn.log.D("[net] broad msg: type= %v, from=%v,to=%v,time=%v, to node: %v", msg.ReqType, msg.From, msg.To, msg.Time, node.Addr())
		if node.Addr() == from {
			continue
		}
		msg.To = node.Addr()
		bn.broadcast(msg)
	}
}

//broadcast broadcast to all neighbours, stop broadcast when msg already broadcast
func (bn *BaseNetwork) broadcast(msg message.Message) {
	node, _ := discover.ParseNode(msg.To)
	if msg.TTL == 0 || bn.localNode.Addr() == node.Addr() {
		return
	} else {
		msg.TTL = msg.TTL - 1
	}
	data, err := msg.Marshal(nil)
	if err != nil {
		bn.log.E("[net] marshal request encountered err:%v", err)
	}
	req := newRequest(BroadcastMessage, bn.localNode.Addr(), data)
	conn, err := bn.dial(msg.To)
	if err != nil {
		bn.log.E("[net] broadcast dial tcp got err:%v", err)
		bn.nodeTable.Delete([]byte(msg.To))
		return
	}
	if er := bn.send(conn, req); er != nil {
		bn.peers.RemoveByNodeStr(msg.To)
	}
}

func (bn *BaseNetwork) dial(nodeStr string) (net.Conn, error) {
	bn.lock.Lock()
	defer bn.lock.Unlock()
	node, _ := discover.ParseNode(nodeStr)
	if bn.localNode.Addr() == node.Addr() {
		return nil, fmt.Errorf("dial local %v", node.Addr())
	}
	peer := bn.peers.Get(node)
	if peer == nil {
		bn.log.D("[net] dial to %v", node.Addr())
		conn, err := net.Dial("tcp", node.Addr())
		if err != nil {
			if conn != nil {
				conn.Close()
			}
			log.Report(&log.MsgNode{SubType: log.Subtypes["MsgNode"][2], Log: node.Addr()})
			bn.log.E("[net] dial tcp %v got err:%v", node.Addr(), err)
			return conn, fmt.Errorf("dial tcp %v got err:%v", node.Addr(), err)
		}
		if conn != nil {
			go bn.receiveLoop(conn)
			peer := newPeer(conn, bn.localNode.Addr(), nodeStr)
			log.Report(&log.MsgNode{SubType: log.Subtypes["MsgNode"][3], Log: node.Addr()})
			log.Report(&log.MsgNode{SubType: log.Subtypes["MsgNode"][4], Log: strconv.Itoa(len(bn.peers.peers))})
			bn.peers.Set(node, peer)
		}
	}

	return bn.peers.Get(node).conn, nil
}

//Send msg to msg.To
func (bn *BaseNetwork) Send(msg message.Message) {
	if msg.To == bn.localNode.Addr() {
		return
	}
	if msg.TTL == 0 {
		return
	} else {
		msg.TTL = msg.TTL - 1
	}
	data, err := msg.Marshal(nil)
	if err != nil {
		bn.log.E("[net] marshal request encountered err:%v", err)
	}
	bn.log.D("[net] send msg: type= %v, from=%v,to=%v,time=%v, to node: %v", msg.ReqType, msg.From, msg.To, msg.Time)
	req := newRequest(Message, bn.localNode.Addr(), data)
	conn, err := bn.dial(msg.To)
	if err != nil {
		bn.nodeTable.Delete([]byte(msg.To))
		bn.log.E("[net] Send, dial tcp got err:%v", err)
		return
	}
	if er := bn.send(conn, req); er != nil {
		bn.peers.RemoveByNodeStr(msg.To)
	}
}

// Close all connection
func (bn *BaseNetwork) Close(port uint16) error {
	if bn.listener != nil {
		bn.listener.Close()
	}
	return nil
}

func (bn *BaseNetwork) send(conn net.Conn, r *Request) error {
	if conn == nil {
		bn.log.E("[net] from %v,send data = %v, conn is nil", bn.localNode.Addr(), r)
		return nil
	}
	pack, err := r.Pack()
	if err != nil {
		bn.log.E("[net] pack data encountered err:%v", err)
		return nil
	}
	_, err = conn.Write(pack)
	if err != nil {
		bn.log.E("[net] conn write got err:%v", err)
	}
	return err
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
			req.handle(bn, conn)
		}
		if err := scanner.Err(); err != nil {
			bn.log.E("[net] invalid data packets: %v", err)
			return
		}
	}
	bn.log.D("[net] recieve loop finish..")
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
	if addrs == "" {
		return
	}
	addrArr := strings.Split(addrs, ",")
	for _, addr := range addrArr {
		node, err := discover.ParseNode(addr)
		if err != nil {
			bn.log.E("failed to ParseNode  %v,err: %v", addr, err)
			continue
		}
		if addr != "" && addr != bn.localNode.Addr() {
			ok, err := bn.nodeTable.Has([]byte(addr))
			if err != nil {
				bn.log.E("failed to nodetable has %v, err: %v", addr, err)
				continue
			}
			if !ok {

				bn.nodeTable.Put([]byte(node.Addr()), common.IntToBytes(2))
			}
		}
	}
	bn.findNeighbours()
	return
}

//nodeCheckLoop inspection Last registration time of node
func (bn *BaseNetwork) nodeCheckLoop() {
	if bn.localNode.TCP == 30304 {
		for {
			iter := bn.nodeTable.NewIterator()
			for iter.Next() {
				k := iter.Key()
				v := common.BytesToInt(iter.Value())
				if v <= 0 {
					bn.log.D("[net] delete node %v, cuz its last register time is %v", string(iter.Key()), common.BytesToInt64(iter.Value()))
					bn.nodeTable.Delete(iter.Key())
					bn.peers.RemoveByNodeStr(string(iter.Key()))
					bn.delNeighbour(string(iter.Key()))
				} else {
					bn.nodeTable.Put(k, common.IntToBytes(v-1))
				}
			}
			time.Sleep(CheckKnownNodeInterval * time.Second)
		}
	}

}

//registerLoop register local address to boot nodes
func (bn *BaseNetwork) registerLoop() {
	for {
		if bn.localNode.TCP != 30304 {
			conn, err := bn.dial(bn.regAddr)
			if err != nil {
				bn.log.E("[net] failed to connect boot node, err:%v", err)
				time.Sleep(CheckKnownNodeInterval * time.Second)
				continue
			}
			bn.log.D("[net] %v request node table from %v", bn.localNode.Addr(), bn.regAddr)
			req := newRequest(ReqNodeTable, bn.localNode.Addr(), nil)
			if er := bn.send(conn, req); er != nil {
				bn.peers.RemoveByNodeStr(bn.regAddr)
			}
		}
		time.Sleep(CheckKnownNodeInterval * time.Second)
	}
}

//findNeighbours find neighbour nodes in the node table
func (bn *BaseNetwork) findNeighbours() {
	nodesStr, _ := bn.AllNodesExcludeAddr(bn.localNode.Addr())
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

func (bn *BaseNetwork) delNeighbour(nodeStr string) {
	bn.lock.Lock()
	defer bn.lock.Unlock()
	delete(bn.neighbours, nodeStr)
}

//Download block by height from which node in NodeHeightMap
func (bn *BaseNetwork) Download(start, end uint64) error {
	bn.lock.Lock()
	for i := start; i <= end; i++ {
		bn.DownloadHeights[i] = 0
	}
	bn.lock.Unlock()

	for retry := 0; retry < MaxDownloadRetry; retry++ {
		wg := sync.WaitGroup{}
		time.Sleep(time.Duration(retry) * time.Second)
		for downloadHeight, retryTimes := range bn.DownloadHeights {
			if retryTimes > MaxDownloadRetry {
				continue
			}
			msg := message.Message{
				Body:    common.Uint64ToBytes(downloadHeight),
				ReqType: int32(ReqDownloadBlock),
				TTL:     MsgMaxTTL,
				From:    bn.localNode.Addr(),
				Time:    time.Now().UnixNano(),
			}
			bn.log.D("[net] download height = %v  nodeMap = %v", downloadHeight, bn.NodeHeightMap)
			bn.lock.Lock()
			bn.DownloadHeights[downloadHeight] = retryTimes + 1
			bn.lock.Unlock()
			wg.Add(1)
			go func() {
				bn.Broadcast(msg)
				wg.Done()
			}()
		}
		wg.Wait()
	}
	return nil
}

//CancelDownload cancel downloading block with height between start and end
func (bn *BaseNetwork) CancelDownload(start, end uint64) error {
	bn.lock.Lock()
	defer bn.lock.Unlock()
	for ; start <= end; start++ {
		delete(bn.DownloadHeights, start)
	}
	return nil
}

//sendTo send request to the address
func (bn *BaseNetwork) sendTo(addr string, req *Request) {
	conn, err := bn.dial(addr)
	if err != nil {
		bn.nodeTable.Delete([]byte(addr))
		bn.log.E("[net] dial tcp got err:%v", err)
		return
	}
	if er := bn.send(conn, req); er != nil {
		bn.peers.RemoveByNodeStr(addr)
	}
}

//SetNodeHeightMap ...
func (bn *BaseNetwork) SetNodeHeightMap(nodeStr string, height uint64) {
	bn.lock.Lock()
	defer bn.lock.Unlock()
	bn.NodeHeightMap[nodeStr] = height
}

//GetNodeHeightMap ...
func (bn *BaseNetwork) GetNodeHeightMap(nodeStr string) uint64 {
	bn.lock.Lock()
	defer bn.lock.Unlock()
	return bn.NodeHeightMap[nodeStr]
}

func randNodeMatchHeight(m map[string]uint64, downloadHeight uint64) (targetNode string) {
	rand.Seed(time.Now().UnixNano())
	matchNum := 1
	for nodeStr, height := range m {
		if height >= downloadHeight {
			randNum := rand.Int31n(int32(matchNum))
			if randNum == 0 {
				targetNode = nodeStr
			}
			matchNum++
		}
	}
	return targetNode
}
