package p2p

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/willf/bloom"
)

const (
	bloomItemCount = 100000
	bloomErrRate   = 0.001

	msgChanSize = 1024
)

type Peer struct {
	peerInfo    *peerstore.PeerInfo
	id          peer.ID
	addr        multiaddr.Multiaddr
	stream      libnet.Stream
	recentMsg   *bloom.BloomFilter
	urgentMsgCh chan *p2pMessage
	normalMsgCh chan *p2pMessage
	quitWriteCh chan struct{}
}

func NewPeer(stream libnet.Stream) *Peer {
	peer := &Peer{
		id:          stream.Conn().RemotePeer(),
		addr:        stream.Conn().RemoteMultiaddr(),
		stream:      stream,
		recentMsg:   bloom.NewWithEstimates(bloomItemCount, bloomErrRate),
		urgentMsgCh: make(chan *p2pMessage, msgChanSize),
		normalMsgCh: make(chan *p2pMessage, msgChanSize),
		quitWriteCh: make(chan struct{}),
	}
	return peer
}

func (p *Peer) write(m *p2pMessage) error {
	deadline := time.Now().Add(time.Duration(len(m.content())/1024/10+1) * time.Second)
	if err := p.stream.SetWriteDeadline(deadline); err != nil {
		// log
		return err
	}
	_, err := p.stream.Write([]byte(*m))
	if err != nil {
		// TODO: log, close
		return err
	}
	// TODO: metrics
	return nil
}

func (p *Peer) writeLoop() {
	for {
		select {
		case <-p.quitWriteCh:
			//log
			return
		case m := <-p.urgentMsgCh:
			p.write(m)
		case m := <-p.normalMsgCh:
			p.write(m)
		}
	}
}

func (p *Peer) readLoop() {
	header := make([]byte, dataBegin)
	for {
		_, err := io.ReadFull(p.stream, header)
		if err != nil {
			// TODO: log, close
			return
		}
		// TODO: check chainID
		length := binary.BigEndian.Uint32(header[dataLengthBegin:dataLengthEnd])
		data := make([]byte, dataBegin+length)
		_, err = io.ReadFull(p.stream, data[dataBegin:])
		if err != nil {
			// TODO: log, close
			return
		}
		copy(data[0:dataBegin], header)
		msg, err := parseP2PMessage(data)
		if err != nil {
			// TODO: log, close
			return
		}
		p.handleMessage(msg)
	}
}

func (p *Peer) SendMessage(msg *p2pMessage, mp MessagePriority) error {
	switch mp {
	case UrgentMessage:
		p.urgentMsgCh <- msg
	default:
		p.normalMsgCh <- msg
	}
	return nil
}

func (p *Peer) handleMessage(msg *p2pMessage) error {
	switch msg.messageType() {
	case Ping:
		fmt.Println("pong")
	}
	return nil
}
