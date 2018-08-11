package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	libnet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/willf/bloom"
)

var (
	ErrStreamCountExceed = errors.New("stream count exceed")
)

const (
	bloomItemCount = 100000
	bloomErrRate   = 0.001

	msgChanSize = 1024

	maxStreamCount = 4
)

type Peer struct {
	id          peer.ID
	addr        multiaddr.Multiaddr
	peerManager *PeerManager

	streams     chan libnet.Stream
	streamCount int
	streamMutex sync.Mutex

	conn        libnet.Conn
	recentMsg   *bloom.BloomFilter
	urgentMsgCh chan *p2pMessage
	normalMsgCh chan *p2pMessage
	quitWriteCh chan struct{}
}

func NewPeer(stream libnet.Stream, pm *PeerManager) *Peer {
	peer := &Peer{
		id:          stream.Conn().RemotePeer(),
		addr:        stream.Conn().RemoteMultiaddr(),
		peerManager: pm,
		conn:        stream.Conn(),
		streams:     make(chan libnet.Stream, maxStreamCount),
		recentMsg:   bloom.NewWithEstimates(bloomItemCount, bloomErrRate),
		urgentMsgCh: make(chan *p2pMessage, msgChanSize),
		normalMsgCh: make(chan *p2pMessage, msgChanSize),
		quitWriteCh: make(chan struct{}),
	}
	peer.AddStream(stream)
	return peer
}

func (p *Peer) Start() {
	p.writeLoop()
}

func (p *Peer) Stop() {
	close(p.quitWriteCh)
	p.conn.Close()
}

func (p *Peer) AddStream(stream libnet.Stream) error {
	p.streamMutex.Lock()
	defer p.streamMutex.Unlock()

	if p.streamCount >= maxStreamCount {
		return ErrStreamCountExceed
	}
	p.streams <- stream
	p.streamCount++
	go p.readLoop(stream)
	return nil
}

func (p *Peer) newStream() (libnet.Stream, error) {
	p.streamMutex.Lock()
	defer p.streamMutex.Unlock()
	if p.streamCount >= maxStreamCount {
		return nil, ErrStreamCountExceed
	}
	stream, err := p.conn.NewStream()
	if err != nil {
		// log
		return nil, err
	}
	p.streamCount++
	return stream, nil
}

func (p *Peer) getStream() (libnet.Stream, error) {
	select {
	case stream := <-p.streams:
		return stream, nil
	default:
		stream, err := p.newStream()
		if err == ErrStreamCountExceed {
			break
		}
		return stream, err
	}
	return <-p.streams, nil
}

func (p *Peer) write(m *p2pMessage) error {
	stream, err := p.getStream()
	if err != nil {
		return err
	}

	deadline := time.Now().Add(time.Duration(len(m.content())/1024/10+1) * time.Second)
	if err = stream.SetWriteDeadline(deadline); err != nil {
		// log
		return err
	}
	_, err = stream.Write(m.content())
	if err != nil {
		// TODO: log
		stream.Reset()
		return err
	}
	p.streams <- stream
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

func (p *Peer) readLoop(stream libnet.Stream) {
	header := make([]byte, dataBegin)
	for {
		_, err := io.ReadFull(stream, header)
		if err != nil {
			// TODO: log
			stream.Reset()
			return
		}
		// TODO: check chainID
		length := binary.BigEndian.Uint32(header[dataLengthBegin:dataLengthEnd])
		data := make([]byte, dataBegin+length)
		_, err = io.ReadFull(stream, data[dataBegin:])
		if err != nil {
			// TODO: log
			stream.Reset()
			return
		}
		copy(data[0:dataBegin], header)
		msg, err := parseP2PMessage(data)
		if err != nil {
			// TODO: log
			stream.Reset()
			return
		}
		p.handleMessage(msg)
	}
}

func (p *Peer) SendMessage(msg *p2pMessage, mp MessagePriority) error {
	// TODO: unblock
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
	default:
		p.peerManager.NotifyMessage(msg, p.id)
	}
	return nil
}
