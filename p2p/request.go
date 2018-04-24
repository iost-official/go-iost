package p2p

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"bytes"

	"net"
)

type ReqType int16

const (
	Message ReqType = iota + 1
	MessageReceived
	BroadcastMessage
	BroadcastMessageReceived
	Ping
	Pong
	ReqNodeTable
	NodeTable
)

// Request 节点之间交换的数据结构
type Request struct {
	Version   [4]byte // 协议版本，暂定iost
	Length    int32   // 数据部分长度
	Timestamp int64   // 时间戳
	Type      ReqType // 传输数据类型
	FromLen   int16
	From      []byte
	Body      []byte
}

var NET_VERSION = [4]byte{'i', 'o', 's', 't'}

func isNetVersionMatch(buf []byte) bool {
	if len(buf) >= len(NET_VERSION) {
		return buf[0] == NET_VERSION[0] &&
			buf[1] == NET_VERSION[1] &&
			buf[2] == NET_VERSION[2] &&
			buf[3] == NET_VERSION[3]
	}
	return false
}

func NewRequest(typ ReqType, from string, data []byte) *Request {
	r := &Request{
		Version:   NET_VERSION,
		Timestamp: time.Now().Unix(),
		Type:      typ,
		FromLen:   int16(len(from)),
		From:      []byte(from),
		Body:      data,
	}
	//len(timestamp) + len(type) + len(fromLen) + len(from) + len(body)
	r.Length = int32(8 + 2 + 2 + len(r.From) + len(data))
	return r
}

func (r *Request) Pack() ([]byte, error) {
	var err error
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, &r.Version)
	err = binary.Write(buf, binary.BigEndian, &r.Length)
	err = binary.Write(buf, binary.BigEndian, &r.Timestamp)
	err = binary.Write(buf, binary.BigEndian, &r.Type)
	err = binary.Write(buf, binary.BigEndian, &r.FromLen)
	err = binary.Write(buf, binary.BigEndian, &r.From)
	err = binary.Write(buf, binary.BigEndian, &r.Body)
	return buf.Bytes(), err
}

func (r *Request) Unpack(reader io.Reader) error {
	var err error
	err = binary.Read(reader, binary.BigEndian, &r.Version)
	err = binary.Read(reader, binary.BigEndian, &r.Length)
	err = binary.Read(reader, binary.BigEndian, &r.Timestamp)
	err = binary.Read(reader, binary.BigEndian, &r.Type)
	err = binary.Read(reader, binary.BigEndian, &r.FromLen)
	r.From = make([]byte, r.FromLen)
	err = binary.Read(reader, binary.BigEndian, &r.From)
	r.Body = make([]byte, r.Length-8-2-2-int32(r.FromLen))
	err = binary.Read(reader, binary.BigEndian, &r.Body)
	return err
}

func (r *Request) String() string {
	return fmt.Sprintf("version:%s length:%d type:%d timestamp:%d from:%s Body:%v",
		r.Version,
		r.Length,
		r.Type,
		r.Timestamp,
		r.From,
		r.Body,
	)
}

//处理接收的数据
//todo 注册， send or 解析 通道map
func (r *Request) handle(s *Server, conn net.Conn) (string, error) {
	s.log.D("handle request = %v", r)
	switch r.Type {
	case Message:
		s.spreadUp(r.Body)
		s.send(conn, nil, MessageReceived)
	case BroadcastMessage:
		s.spreadUp(r.Body)
		s.broadcast(r)
		s.send(conn, nil, MessageReceived)
	case MessageReceived:
	case BroadcastMessageReceived:
	case Ping:
		// a downstream node sends its address to its seed node(our node)
		addr := string(r.Body)
		s.setPeer(addr, conn)
		s.putNode(addr)
		//返回pong
		s.send(conn, nil, Pong)
		return addr, nil
	case Pong:
		s.pinged = true
	case NodeTable: //got nodeTable and save
		s.appendNodeTable(string(r.Body))
	case ReqNodeTable: //request for nodeTable
		addrs, err := s.allNodesExcludeAddr(string(r.From))
		if err != nil {
			s.log.E("failed to nodetable ", err)
		}
		s.send(conn, []byte(addrs), NodeTable)
	default:
		s.log.E("wrong request :", r)
	}
	return "", nil
}
