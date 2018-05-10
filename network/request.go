package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"bytes"

	"net"

	"strings"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/message"
)

type NetReqType int16

const (
	Message NetReqType = iota + 1
	MessageReceived
	BroadcastMessage
	BroadcastMessageReceived
	Ping
	Pong
	ReqNodeTable
	NodeTable
)

// Request data structure exchanged by nodes
type Request struct {
	Version   [4]byte
	Length    int32 // length of request
	Timestamp int64
	Type      NetReqType
	FromLen   int16
	From      []byte
	Priority  int8
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

func newRequest(typ NetReqType, from string, data []byte) *Request {
	r := &Request{
		Version:   NET_VERSION,
		Timestamp: time.Now().UnixNano(),
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
	return fmt.Sprintf("version:%s length:%d type:%d timestamp:%d from:%s priority %v Body:%v",
		r.Version,
		r.Length,
		r.Type,
		r.Timestamp,
		r.From,
		r.Priority,
		r.Body,
	)
}

func (r *Request) handle(s *Server, conn net.Conn) (string, error) {
	s.log.D("handle request = %v", r)
	switch r.Type {
	case Message:
		s.spreadUp(r.Body)
		req := newRequest(MessageReceived, s.ListenAddr, common.Int64ToBytes(r.Timestamp))
		s.send(conn, req)
	case BroadcastMessage:
		s.spreadUp(r.Body)
		s.broadcast(*r)
		req := newRequest(BroadcastMessageReceived, s.ListenAddr, common.Int64ToBytes(r.Timestamp))
		s.send(conn, req)
	case MessageReceived:
		s.log.D("MessageReceived: %v", common.BytesToInt64(r.Body))
	case BroadcastMessageReceived:
		s.log.D("BroadcastMessageReceived: %v", common.BytesToInt64(r.Body))
	case Ping:
		// a downstream node sends its address to its seed node(our node)
		addr := string(r.Body)
		s.addPeer(addr, conn)
		s.putNode(addr)
		//return pong
		req := newRequest(Pong, s.ListenAddr, nil)
		s.send(conn, req)
		return addr, nil
	case Pong:
		s.pinged = true
	case NodeTable: //got nodeTable and save
		s.putNode(string(r.Body))
	case ReqNodeTable: //request for nodeTable
		addrs, err := s.AllNodesExcludeAddr(string(r.From))
		if err != nil {
			s.log.E("failed to nodetable ", err)
		}
		req := newRequest(NodeTable, s.ListenAddr, []byte(addrs))
		s.send(conn, req)
	default:
		s.log.E("wrong request :", r)
	}
	return "", nil
}

func (r *Request) response(base *BaseNetwork, conn net.Conn) {
	base.log.D("response request = %v", r)
	switch r.Type {
	case Message:
		appReq := &message.Message{}
		if _, err := appReq.Unmarshal(r.Body); err == nil {
			base.RecvCh <- *appReq
		}
		base.send(conn, newRequest(MessageReceived, base.localNode.String(), common.Int64ToBytes(r.Timestamp)))
	case MessageReceived:
		base.log.D("MessageReceived: %v", common.BytesToInt64(r.Body))
		r.broadcastMsgHandle(base)
	case BroadcastMessage:
		appReq := &message.Message{}
		if _, err := appReq.Unmarshal(r.Body); err == nil {
			base.RecvCh <- *appReq
			base.Broadcast(*appReq)
		}
	case BroadcastMessageReceived:
	//request for nodeTable
	case ReqNodeTable:
		base.putNode(string(r.From))
		addrs, err := base.AllNodesExcludeAddr(string(r.From))
		if err != nil {
			base.log.E("failed to nodetable ", err)
		}
		req := newRequest(NodeTable, base.localNode.String(), []byte(strings.Join(addrs, ",")))
		base.send(conn, req)
	//got nodeTable and save
	case NodeTable:
		base.putNode(string(r.Body))
	default:
		base.log.E("wrong request :", r)
	}
}

//handle broadcast node's height
func (r *Request) broadcastMsgHandle(net *BaseNetwork) {
	msg := &message.Message{}
	if _, err := msg.Unmarshal(r.Body); err == nil {
		switch msg.ReqType {
		case int32(ReqBlockHeight):
			net.SetNodeHeightMap(string(r.From), common.BytesToUint64(msg.Body))
		default:

		}
	}
}
