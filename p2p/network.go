package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	"strings"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core"
	"github.com/iost-official/prototype/iostdb"
)

type RequestHead struct {
	Length uint32 // Request的长度信息
}

const HEADLENGTH = 4

type Response struct {
	From        string
	To          string
	Code        int // http-like的状态码和描述
	Description string
}

//Network 最基本网络的模块API，之后gossip协议，虚拟的网络延迟都可以在模块内部实现
type Network interface {
	Broadcast(req core.Request)
	Send(req core.Request)
	Listen(port uint16) (<-chan core.Request, error)
	Close(port uint16) error
}

type NaiveNetwork struct {
	db     *iostdb.LDBDatabase //database of known nodes
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
	db, err := iostdb.NewLDBDatabase(dirname, 0, 0)
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
func (nn *NaiveNetwork) Broadcast(req core.Request) error {
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

func (nn *NaiveNetwork) Send(req core.Request) {
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

func (nn *NaiveNetwork) Listen(port uint16) (<-chan core.Request, error) {
	var err error
	nn.listen, err = net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		return nil, fmt.Errorf("Error listening: %v", err.Error())
	}
	fmt.Println("Listening on " + ":" + strconv.Itoa(int(port)))
	req := make(chan core.Request)

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
					var received core.Request
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

func reqToBytes(req core.Request) ([]byte, []byte, error) {
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

const (
	FINDNODEINTERVAL = 10 * time.Second
)

type NetworkImpl struct {
	db     *iostdb.LDBDatabase //database of known nodes
	listen net.Listener
	conn   net.Conn
	done   bool
}

//建立tcp连接，读取数据，
func (n *NetworkImpl) Listen(tcpAddr string) (err error) {
	// 轮询接收结果，轮询发送请求给随机节点
	n.listen, err = net.Listen("tcp", tcpAddr)
	if err != nil {
		return
	}
	n.conn, err = n.listen.Accept()
	if err != nil {
		n.conn = nil
		return
	}
	go n.readLoop()
	go n.findNodeTableLoop()
	return
}

func (n *NetworkImpl) readLoop() {
	//临时缓冲区，用来存储被截断的数据
	tmpBuffer := make([]byte, 0)

	readerChannel := make(chan []byte)
	go n.reader(readerChannel)

	buffer := make([]byte, 1024)
	for {
		n, err := n.conn.Read(buffer)
		if err != nil {
			return
		}
		tmpBuffer = Unpack(append(tmpBuffer, buffer[:n]...), readerChannel)
	}
}

func (n *NetworkImpl) reader(readerCh chan []byte) {
	for {
		select {
		case data := <-readerCh:
			var received Msg
			_, err := received.Unmarshal(data)
			fmt.Errorf("unmarshal got err:%+v", err)
			switch received.Code {
			case uint64(findnodePacket):
				n.conn.Write(n.DBBytes())
			case uint64(storenodePacket):
				n.appendDB(received)
			default:
				//req <- received //todo:返回给上层
			}
		case <-time.After(100 * time.Second):
			n.conn.Close()
		}
	}
}

func (n *NetworkImpl) findNodeTableLoop() {
	for {
		time.Sleep(FINDNODEINTERVAL)
		if n.conn == nil {
			continue
		}
		m := Msg{Code: findnodePacket, Size: 0}
		n.Send(m)
	}
}

var addrSeparator = []byte("_")

func (n *NetworkImpl) DBBytes() []byte {
	addrs := make([]byte, 0)
	iter := n.db.NewIterator()
	for iter.Next() {
		addr, _ := n.db.Get([]byte(string(iter.Key())))
		addrs = append(addrs, addr...)
		addrs = append(addrs, addrSeparator...)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		fmt.Errorf("node table iterator got err : %v", err)
	}
	return addrs
}

func (n *NetworkImpl) appendDB(msg Msg) {
	data, err := ioutil.ReadAll(msg.Payload)
	if err != nil {
		fmt.Errorf("io read ecounter %+v", err)
		return
	}
	addrs := strings.Split(string(data), string(addrSeparator))
	for _, addr := range addrs {
		k := common.Sha256([]byte(addr))
		if ok, _ := n.db.Has([]byte(k)); ok {
			n.db.Put([]byte(k), []byte(addr))
		}
	}
}

func (nn *NetworkImpl) Send(msg Msg) {
	if nn.conn == nil {
		fmt.Errorf("no connect in network")
		return
	}
	m := Msg{Code: findnodePacket, Size: 0}
	buf, _ := m.Marshal(nil)
	if _, err := nn.conn.Write(buf); err != nil {
		fmt.Errorf("sending msg encounter err :%v\n", err.Error())
	}
	return
}
