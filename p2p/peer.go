package p2p

import (
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/iost-official/go-iost/ilog"

	libnet "github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/willf/bloom"
)

// errors
var (
	ErrStreamCountExceed  = errors.New("stream count exceed")
	ErrMessageChannelFull = errors.New("message channel is full")
	ErrDuplicateMessage   = errors.New("reduplicate message")
)

const (
	bloomMaxItemCount = 100000
	bloomErrRate      = 0.001

	msgChanSize = 1024
)

// Peer represents a neighbor which we connect directily.
//
// Peer's jobs are:
//   * managing streams which are responsible for sending and reading messages.
//   * recording messages we have sent and received so as to reduce redundant message in network.
//   * maintaning a priority queue of message to be sending.
type Peer struct {
	id          peer.ID
	addr        multiaddr.Multiaddr
	conn        libnet.Conn
	peerManager *PeerManager

	stream libnet.Stream

	recentMsg      *bloom.BloomFilter
	bloomMutex     sync.Mutex
	bloomItemCount int

	urgentMsgCh chan *p2pMessage
	normalMsgCh chan *p2pMessage

	direction connDirection

	quitWriteCh chan struct{}
	once        sync.Once
}

// NewPeer returns a new instance of Peer struct.
func NewPeer(stream libnet.Stream, pm *PeerManager, direction connDirection) *Peer {
	peer := &Peer{
		id:          stream.Conn().RemotePeer(),
		addr:        stream.Conn().RemoteMultiaddr(),
		conn:        stream.Conn(),
		stream:      stream,
		peerManager: pm,
		recentMsg:   bloom.NewWithEstimates(bloomMaxItemCount, bloomErrRate),
		urgentMsgCh: make(chan *p2pMessage, msgChanSize),
		normalMsgCh: make(chan *p2pMessage, msgChanSize),
		quitWriteCh: make(chan struct{}),
		direction:   direction,
	}
	go peer.readLoop(stream)
	return peer
}

// ID return the net id.
func (p *Peer) ID() string {
	return p.id.Pretty()
}

// Addr return the address.
func (p *Peer) Addr() string {
	return p.addr.String()
}

// Start starts peer's loop.
func (p *Peer) Start() {
	ilog.Infof("peer is started. id=%s", p.id.Pretty())

	go p.writeLoop()
}

// Stop stops peer's loop and cuts off the TCP connection.
func (p *Peer) Stop() {
	ilog.Infof("peer is stopped. id=%s", p.id.Pretty())

	p.once.Do(func() {
		close(p.quitWriteCh)
	})
	p.conn.Close()
}

// AddStream tries to add a Stream in stream pool.
func (p *Peer) AddStream(stream libnet.Stream) error {
	/*  if err := p.streamPool.Put(stream); err != nil { */
	// return err
	/* } */
	go p.readLoop(stream)
	return nil
}

func (p *Peer) write(m *p2pMessage) error {
	// s, err := p.streamPool.Get()
	// if getStream fails, the TCP connection may be broken and we should stop the peer.
	/* if err != nil { */
	// ilog.Errorf("get stream fails. err=%v", err)
	// p.peerManager.RemoveNeighbor(p.id)
	// return err
	// }
	/* stream := s.(libnet.Stream) */

	// 5 kB/s
	deadline := time.Now().Add(time.Duration(len(m.content())/1024/5+1) * time.Second)
	if err := p.stream.SetWriteDeadline(deadline); err != nil {
		ilog.Warnf("set write deadline failed. err=%v", err)
		p.stream.Close()
		return err
	}
	_, err := p.stream.Write(m.content())
	if err != nil {
		ilog.Warnf("write message failed. err=%v", err)
		// p.stream.Close()
		p.peerManager.RemoveNeighbor(p.id)
		return err
	}
	tagkv := map[string]string{"mtype": m.messageType().String()}
	byteOutCounter.Add(float64(len(m.content())), tagkv)
	packetOutCounter.Add(1, tagkv)

	return nil
}

func (p *Peer) writeLoop() {
	for {
		select {
		case <-p.quitWriteCh:
			ilog.Infof("peer is stopped. pid=%v, addr=%v", p.id.Pretty(), p.addr)
			return
		case um := <-p.urgentMsgCh:
			p.write(um)
		case nm := <-p.normalMsgCh:
			for done := false; !done; {
				select {
				case <-p.quitWriteCh:
					ilog.Infof("peer is stopped. pid=%v, addr=%v", p.id.Pretty(), p.addr)
					return
				case um := <-p.urgentMsgCh:
					p.write(um)
				default:
					done = true
				}
			}
			p.write(nm)
		}
	}
}

func (p *Peer) readLoop(stream libnet.Stream) {
	header := make([]byte, dataBegin)
	for {
		_, err := io.ReadFull(stream, header)
		if err != nil {
			ilog.Warnf("read header failed. err=%v", err)
			return
		}
		chainID := binary.BigEndian.Uint32(header[chainIDBegin:chainIDEnd])
		if chainID != p.peerManager.config.ChainID {
			ilog.Warnf("mismatched chainID. chainID=%d", chainID)
			return
		}
		length := binary.BigEndian.Uint32(header[dataLengthBegin:dataLengthEnd])
		data := make([]byte, dataBegin+length)
		_, err = io.ReadFull(stream, data[dataBegin:])
		if err != nil {
			ilog.Warnf("read message failed. err=%v", err)
			return
		}
		copy(data[0:dataBegin], header)
		msg, err := parseP2PMessage(data)
		if err != nil {
			ilog.Errorf("parse p2pmessage failed. err=%v", err)
			return
		}
		tagkv := map[string]string{"mtype": msg.messageType().String()}
		byteInCounter.Add(float64(len(msg.content())), tagkv)
		packetInCounter.Add(1, tagkv)
		p.handleMessage(msg)
	}
}

// SendMessage puts message into the corresponding channel.
func (p *Peer) SendMessage(msg *p2pMessage, mp MessagePriority, deduplicate, async bool) error {
	if deduplicate && msg.needDedup() {
		if p.hasMessage(msg) {
			// ilog.Debug("ignore reduplicate message")
			return ErrDuplicateMessage
		}
	}
	if !async {
		if msg.needDedup() {
			p.recordMessage(msg)
		}
		return p.write(msg)
	}
	ch := p.urgentMsgCh
	if mp == NormalMessage {
		ch = p.normalMsgCh
	}
	select {
	case ch <- msg:
	default:
		ilog.Errorf("sending message failed. channel is full. messagePriority=%d", mp)
		return ErrMessageChannelFull
	}
	if msg.needDedup() {
		p.recordMessage(msg)
	}
	return nil
}

func (p *Peer) handleMessage(msg *p2pMessage) error {
	if msg.needDedup() {
		p.recordMessage(msg)
	}
	p.peerManager.HandleMessage(msg, p.id)
	return nil
}

func (p *Peer) recordMessage(msg *p2pMessage) {
	p.bloomMutex.Lock()
	defer p.bloomMutex.Unlock()

	if p.bloomItemCount >= bloomMaxItemCount {
		p.recentMsg = bloom.NewWithEstimates(bloomMaxItemCount, bloomErrRate)
		p.bloomItemCount = 0
	}

	p.recentMsg.Add(msg.content())
	p.bloomItemCount++
}

func (p *Peer) hasMessage(msg *p2pMessage) bool {
	p.bloomMutex.Lock()
	defer p.bloomMutex.Unlock()

	return p.recentMsg.Test(msg.content())
}
