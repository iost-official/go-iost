package p2p

import (
	"net"
	"fmt"
	"encoding/binary"
	"unsafe"
	"bytes"
)

type Request struct {
	Time    int64  // 发送时的时间戳
	From    string // From To是钱包地址的base58编码字符串（就是Member.ID，下同）
	To      string
	ReqType int // 此request的类型码，通过类型可以确定body的格式以方便解码body
	Body    []byte
}

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

func (network *NaiveNetwork) Close(port uint16) {
	network.done = true
	network.listen.Close()
}

func (network *NaiveNetwork) Send(req Request) {
	length := unsafe.Sizeof(req)
	int32buf := new(bytes.Buffer)
	binary.Write(int32buf, binary.BigEndian, length)
	for _, addr := range network.peerList {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Println("Error dialing to ", addr, err.Error())
		}
		if _, err = conn.Write(int32buf.Bytes()); err != nil {
			fmt.Println("Error sending request head:", err.Error())
		}
		buf := *(*[length]byte)(unsafe.Pointer(&req))
		if _, err = conn.Write(buf[:]); err != nil {
			fmt.Println("Error sending request body:", err.Error())
		}
		conn.Close()
	}
}

func (network *NaiveNetwork) Listen(port uint16) (chan<- Request, error) {
	var err error
	network.listen, err = net.Listen("tcp", ":"+string(port))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return nil, err
	}
	fmt.Println("Listening on " + ":" + string(port))

	req := make(chan Request)
	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := network.listen.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				if network.done {
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

				req <- *(*Request)(unsafe.Pointer(&_buf))
				// Send a response back to person contacting us.
				conn.Write([]byte("Message received."))
			}(conn)
		}

	}()
	return req, nil
}
