package p2p

import (
	"net"
	"fmt"
	"encoding/binary"
	"bytes"
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

// 最基本网络的模块API，之后gossip协议，虚拟的网络延迟都可以在模块内部实现
type Network interface {
	Send(req Request)
	Listen(port uint16) (chan<- Request, error)
	Close(port uint16) error
}

type NaiveNetwork struct {
	peerList []string
	listen   net.Listener
	done     bool
}

func NewNaiveNetwork() *NaiveNetwork {
	nn := &NaiveNetwork{
		peerList: []string{"1.1.1.1", "2.2.2.2"},
		listen:   nil,
		done:     false,
	}
	return nn
}

func (nn *NaiveNetwork) Close(port uint16) error {
	port = 3 // 避免出现unused variable
	nn.done = true
	return nn.listen.Close()
}

func (nn *NaiveNetwork) Send(req Request) {
	buf, err := req.Marshal(nil)
	if err != nil {
		fmt.Println("Error marshal body:", err.Error())
	}
	length := len(buf)
	int32buf := new(bytes.Buffer)
	binary.Write(int32buf, binary.BigEndian, length)
	for _, addr := range nn.peerList {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Println("Error dialing to ", addr, err.Error())
		}
		if _, err = conn.Write(int32buf.Bytes()); err != nil {
			fmt.Println("Error sending request head:", err.Error())
		}
		if _, err = conn.Write(buf[:]); err != nil {
			fmt.Println("Error sending request body:", err.Error())
		}
		conn.Close()
	}
}

func (nn *NaiveNetwork) Listen(port uint16) (chan<- Request, error) {
	var err error
	nn.listen, err = net.Listen("tcp", ":"+string(port))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return nil, err
	}
	fmt.Println("Listening on " + ":" + string(port))

	req := make(chan Request)
	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := nn.listen.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				if nn.done {
					return
				}
				continue
			}
			// Handle connections in a new goroutine.
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
				var received Request
				received.Unmarshal(_buf)
				req <- received
				// Send a response back to person contacting us.
				//conn.Write([]byte("Message received."))
			}(conn)
		}

	}()
	return req, nil
}
