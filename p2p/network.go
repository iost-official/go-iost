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

	"github.com/iost-official/prototype/iostdb"
	"github.com/iost-official/prototype/core"
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
	Listen(port uint16) (<-chan core.Request, error)
	Close(port uint16) error
}

type NaiveNetwork struct {
	peerList *iostdb.LDBDatabase
	listen   net.Listener
	conn     net.Conn
	done     bool
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
		peerList: db,
		listen:   nil,
		done:     false,
	}
	for i:= 1; i <= n; i++ {
		nn.peerList.Put([]byte(string(i)), []byte("127.0.0.1:" + strconv.Itoa(11036 + i)))
	}
	return nn,nil
}

func (nn *NaiveNetwork) Close(port uint16) error {
	port = 3 // 避免出现unused variable
	nn.done = true
	return nn.listen.Close()
}

func (nn *NaiveNetwork) Broadcast(req core.Request) error {
	iter := nn.peerList.NewIterator()
	for iter.Next() {
		addr, _ := nn.peerList.Get([]byte(string(iter.Key())))
		conn, err := net.Dial("tcp", string(addr))
		if err != nil {
			fmt.Errorf("dialing to %v encounter err : %v\n", addr, err.Error())
			continue
		}
		nn.conn = conn
		go nn.Send(req)
	}
	return nil
}

func (nn *NaiveNetwork) Send(req core.Request) {
	if nn.conn == nil {
		fmt.Errorf("no connect in network")
		return
	}
	defer nn.conn.Close()

	reqHeadBytes, reqBodyBytes, err := reqToBytes(req)
	if err !=nil {
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
					//fmt.Printf("got %+v %+v\n", received, port)
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

func reqToBytes(req core.Request) ( []byte,  []byte, error) {
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

