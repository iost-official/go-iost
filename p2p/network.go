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
	Send(req core.Request)
	Listen(port uint16) (<-chan core.Request, error)
	Close(port uint16) error
}

type NaiveNetwork struct {
	peerList *iostdb.LDBDatabase
	listen   net.Listener
	done     bool
}

func NewNaiveNetwork() *NaiveNetwork {
	dirname, _ := ioutil.TempDir(os.TempDir(), "p2p_test_")
	db, _ := iostdb.NewLDBDatabase(dirname, 0, 0)
	nn := &NaiveNetwork{
		//peerList: []string{"1.1.1.1", "2.2.2.2"},
		peerList: db,
		listen:   nil,
		done:     false,
	}
	nn.peerList.Put([]byte("1"), []byte("127.0.0.1:11037"))
	nn.peerList.Put([]byte("2"), []byte("127.0.0.1:11038"))
	nn.peerList.Put([]byte("3"), []byte("127.0.0.1:11039"))
	return nn
}

func (nn *NaiveNetwork) Close(port uint16) error {
	port = 3 // 避免出现unused variable
	nn.done = true
	return nn.listen.Close()
}

func (nn *NaiveNetwork) Send(req core.Request) error {
	buf, err := req.Marshal(nil)
	if err != nil {
		return err
	}

	length := int32(len(buf))
	int32buf := new(bytes.Buffer)

	if err = binary.Write(int32buf, binary.BigEndian, length); err != nil {
		return err
	}
	for i := 1; i < 4; i++ {
		addr, _ := nn.peerList.Get([]byte(strconv.Itoa(i)))
		conn, err := net.Dial("tcp", string(addr))
		if err != nil {
			fmt.Errorf("dialing to %v encounter err : %v\n", addr, err.Error())
			continue
		}
		defer conn.Close()
		if _, err = conn.Write(int32buf.Bytes()); err != nil {
			fmt.Errorf("sending request head encounter err :%v\n", err.Error())
			continue
		}
		if _, err = conn.Write(buf[:]); err != nil {
			fmt.Errorf("sending request body encounter err : %v\n", err.Error())
			continue
		}
	}
	return nil
}

func (nn *NaiveNetwork) Listen(port uint16) (<-chan core.Request, error) {
	var err error
	nn.listen, err = net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return nil, err
	}
	fmt.Println("Listening on " + ":" + strconv.Itoa(int(port)))
	req := make(chan core.Request)

	conn := make(chan net.Conn)

	// For every listener spawn the following routine
	go func(l net.Listener) {
		for {
			//fmt.Println("new conn1", port)
			c, err := l.Accept()
			//fmt.Println("new conn", port)
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
						fmt.Println("Error reading request head:", err.Error())
					}
					length := binary.BigEndian.Uint32(buf)
					_buf := make([]byte, length)
					_, err = conn.Read(_buf)

					if err != nil {
						fmt.Println("Error reading request body:", err.Error())
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

