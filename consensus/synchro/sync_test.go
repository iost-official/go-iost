package synchro

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/chainbase"
	"github.com/iost-official/go-iost/consensus/synchro/pb"
	"github.com/iost-official/go-iost/core/block"
	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/blockcache/mock"
	core_mock "github.com/iost-official/go-iost/core/mocks"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	p2p_mock "github.com/iost-official/go-iost/p2p/mocks"
)

type registerInfo struct {
	peerID p2p.PeerID
	ch     chan p2p.IncomingMessage
	typs   []p2p.MessageType
}
type msgCenter struct {
	r []registerInfo
}

func (m *msgCenter) register(id p2p.PeerID, typs ...p2p.MessageType) chan p2p.IncomingMessage {
	ch := make(chan p2p.IncomingMessage, 4096)
	m.r = append(m.r, registerInfo{id, ch, typs})
	return ch
}
func (m *msgCenter) broadcast(from p2p.PeerID, msg []byte, typ p2p.MessageType) {
	for _, r := range m.r {
		if r.peerID == from {
			continue
		}
		for _, t := range r.typs {
			if t == typ {
				r.ch <- *p2p.NewIncomingMessage(from, msg, typ)
				break
			}
		}
	}
}
func (m *msgCenter) sendToPeer(from p2p.PeerID, to p2p.PeerID, msg []byte, typ p2p.MessageType) {
	for _, r := range m.r {
		if r.peerID == to {
			for _, t := range r.typs {
				if t == typ {
					r.ch <- *p2p.NewIncomingMessage(from, msg, typ)
					break
				}
			}
		}
	}
}

type dataCenter struct {
	chains     [][]int
	blocks     map[int]*blockcache.BlockCacheNode
	blockHashs map[string]int
}

// Simply assume the height of block with id x is x%magicHeight :D
const magicHeight = 10000

// addChain(5,10010,20015) will add blocks 0,1,...,5,10006,10007,...,10010,20011,20012,...,20015.
func (d *dataCenter) addChain(chain ...int) {
	d.chains = append(d.chains, chain)
	parent := -1
	start := 0
	for idx, x := range chain {
		if idx == 0 {
			if x >= magicHeight {
				panic(fmt.Sprintf("please specify the blocks between 0 and %v", x))
			}
		} else {
			if x%magicHeight <= chain[idx-1]%magicHeight {
				panic(fmt.Sprintf("heigh of block %v is not greater than block %v", x, chain[idx-1]))
			}
			start = x - x%magicHeight + chain[idx-1]%magicHeight + 1
		}
		for i := start; i <= x; i++ {
			var parentHash []byte
			if b, ok := d.blocks[parent]; ok {
				parentHash = b.HeadHash()
			}
			newBlock := blockcache.BlockCacheNode{
				Block: &block.Block{
					Head: &block.BlockHead{
						ParentHash: parentHash,
						Number:     int64(i % magicHeight),
						Info:       []byte(strconv.Itoa(i)),
					},
					Sign: &crypto.Signature{},
				},
			}
			newBlock.Block.CalculateHeadHash()
			if b, ok := d.blocks[i]; ok {
				if string(newBlock.HeadHash()) != string(b.HeadHash()) {
					// The only way to make 2 blocks with same id have different hashs is with different parents.
					panic(fmt.Sprintf("invalid parent %v of block %v, should be %v", parent, i, d.getParentID(i)))
				}
			} else {
				log.Println("Data center add new block", i, "with parent", parent)
				d.blocks[i] = &newBlock
				d.blockHashs[string(newBlock.HeadHash())] = i
			}
			parent = i
		}
	}
}
func (d *dataCenter) getParentID(x int) int {
	return d.blockHashs[string(d.blocks[x].Block.Head.ParentHash)]
}
func (d *dataCenter) getLongest(blocks *map[int]bool) (int, *blockcache.BlockCacheNode) {
	longest := 0
	longestChain := 0
	for chainIdx, chain := range d.chains {
		start := 0
		head := 0
		for idx, x := range chain {
			if idx > 0 {
				start = x - x%magicHeight + chain[idx-1]%magicHeight + 1
			}
			for i := start; i <= x; i++ {
				if (*blocks)[i] {
					head = i
				} else {
					break
				}
			}
			if head != x {
				break
			}
		}
		if head%magicHeight > longest%magicHeight {
			longest = head
			longestChain = chainIdx
		}
	}
	return longestChain, d.blocks[longest]
}
func (d *dataCenter) getBlockByNumber(chainIdx int, num int) *block.Block {
	chain := d.chains[chainIdx]
	for _, x := range chain {
		if x%magicHeight >= num {
			return d.blocks[x-x%magicHeight+num].Block
		}
	}
	panic("unreachable")
	return nil
}

func getID(b *block.Block) int {
	x, _ := strconv.Atoi(string(b.Head.Info))
	return x
}

type peer struct {
	mCenter *msgCenter
	dCenter *dataCenter

	id     p2p.PeerID
	blocks map[int]bool
	mu     sync.RWMutex
	quitCh chan struct{}

	sync *Sync

	head         *blockcache.BlockCacheNode
	lib          *blockcache.BlockCacheNode
	longestChain int
}

func newPeer(m *msgCenter, d *dataCenter, id string) *peer {
	return &peer{
		mCenter: m,
		dCenter: d,
		id:      p2p.PeerID(id),
		blocks:  make(map[int]bool),
		quitCh:  make(chan struct{}),
	}
}

func (p *peer) close() {
	p.sync.Close()
	close(p.quitCh)
	log.Println(string(p.id), "closed")
}

func (p *peer) addBlocks(start int, end int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i := start; i <= end; i++ {
		if _, ok := p.dCenter.blocks[i]; !ok {
			panic(fmt.Sprintf("node %v does not exist in data center", i))
		}
		if i > start {
			if p.dCenter.getParentID(i) != i-1 {
				panic(fmt.Sprintf("parent of node %v should be %v instead of %v", i, p.dCenter.getParentID(i), i-1))
			}
		}
		if !p.blocks[i] {
			log.Println(string(p.id), "add block", i)
			p.blocks[i] = true
		}
	}
	p.longestChain, p.head = p.dCenter.getLongest(&p.blocks)
	log.Println(string(p.id), "has head", getID(p.head.Block))
}

func (p *peer) setLib(lib int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.dCenter.blocks[lib]; !ok {
		panic(fmt.Sprintf("node %v does not exist in data center", lib))
	}
	if !p.blocks[lib] {
		panic(fmt.Sprintf("%v does not have the block %v", string(p.id), lib))
	}
	p.lib = p.dCenter.blocks[lib]
	log.Println(string(p.id), "has lib", lib)
}

// Mock funcs.
func (p *peer) p2pRegister(id string, typs ...p2p.MessageType) chan p2p.IncomingMessage {
	log.Println(string(p.id), "register for", id)
	return p.mCenter.register(p.id, typs...)
}
func (p *peer) p2pBroadcast(msg []byte, typ p2p.MessageType, _ p2p.MessagePriority) {
	info := typ.String()
	switch typ {
	case p2p.SyncHeight:
		resp := &msgpb.SyncHeight{}
		proto.Unmarshal(msg, resp)
		info += fmt.Sprintf(" %v", resp.Height)
	case p2p.SyncBlockHashRequest:
		req := &msgpb.BlockHashQuery{}
		proto.Unmarshal(msg, req)
		info += fmt.Sprintf(" %v~%v", req.Start, req.End)
	case p2p.NewBlockHash:
		resp := &msgpb.BlockInfo{}
		proto.Unmarshal(msg, resp)
		info += fmt.Sprintf(" %v", p.dCenter.blockHashs[string(resp.Hash)])
	}
	log.Println(string(p.id), "broadcast a msg", info)
	p.mCenter.broadcast(p.id, msg, typ)
}
func (p *peer) p2pSendToPeer(to p2p.PeerID, msg []byte, typ p2p.MessageType, _ p2p.MessagePriority) {
	info := typ.String()
	switch typ {
	case p2p.SyncBlockHashResponse:
		resp := &msgpb.BlockHashResponse{}
		proto.Unmarshal(msg, resp)
		info += " ["
		for _, b := range resp.BlockInfos {
			info += fmt.Sprintf("%v, ", p.dCenter.blockHashs[string(b.Hash)])
		}
		info += "]"
	case p2p.SyncBlockRequest:
		req := &msgpb.BlockInfo{}
		proto.Unmarshal(msg, req)
		info += fmt.Sprintf(" %v", p.dCenter.blockHashs[string(req.Hash)])
	case p2p.SyncBlockResponse:
		b := &block.Block{}
		b.Decode(msg)
		info += fmt.Sprintf(" %v", getID(b))
	}
	log.Println(string(p.id), "send a msg", info, "to", string(to))
	p.mCenter.sendToPeer(p.id, to, msg, typ)
}
func (p *peer) bcacheHead() *blockcache.BlockCacheNode {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.head
}
func (p *peer) bcacheLinkedRoot() *blockcache.BlockCacheNode {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lib
}
func (p *peer) bcacheFind(hash []byte) (*blockcache.BlockCacheNode, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	x := p.dCenter.blockHashs[string(hash)]
	if !p.blocks[x] {
		return nil, errors.New("block not found")
	} else {
		return p.dCenter.blocks[x], nil
	}
}
func (p *peer) bcacheGetBlockByHash(hash []byte) (*block.Block, error) {
	b, err := p.bcacheFind(hash)
	if err != nil {
		return nil, err
	}
	return b.Block, nil
}
func (p *peer) bcacheGetBlockByNumber(num int64) (*block.Block, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	// Same logic with blockcache.BlockCacheImpl.GetBlockByNumber().
	if num < p.lib.Head.Number || num > p.head.Head.Number {
		return nil, errors.New("block not found")
	}
	return p.dCenter.getBlockByNumber(p.longestChain, int(num)), nil
}
func (p *peer) pobDoVerifyBlock(b *block.Block) {
	// Brief logic of pob.PoB.doVerifyBlock().
	x := getID(b)
	log.Println(string(p.id), "gets new block", x)
	p.addBlocks(x, x)
	if !p.sync.IsCatchingUp() {
		p.sync.BroadcastBlockInfo(b)
	}
}

func (p *peer) NewSync(ctrl *gomock.Controller) {
	p2pService := p2p_mock.NewMockService(ctrl)
	p2pService.EXPECT().Register(gomock.Any(), gomock.Any()).DoAndReturn(p.p2pRegister).AnyTimes()
	p2pService.EXPECT().Broadcast(gomock.Any(), gomock.Any(), gomock.Any()).Do(p.p2pBroadcast).AnyTimes()
	p2pService.EXPECT().SendToPeer(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(p.p2pSendToPeer).AnyTimes()

	bCache := mock.NewMockBlockCache(ctrl)
	bCache.EXPECT().Head().DoAndReturn(p.bcacheHead).AnyTimes()
	bCache.EXPECT().LinkedRoot().DoAndReturn(p.bcacheLinkedRoot).AnyTimes()
	bCache.EXPECT().GetBlockByHash(gomock.Any()).DoAndReturn(p.bcacheGetBlockByHash).AnyTimes()
	bCache.EXPECT().GetBlockByNumber(gomock.Any()).DoAndReturn(p.bcacheGetBlockByNumber).AnyTimes()

	// bChain will be used if failed to do bCache.GetBlockByNumber() in synchro.requestHandler.getBlockHashResponse().
	bChain := core_mock.NewMockChain(ctrl)
	bChain.EXPECT().GetHashByNumber(gomock.Any()).Return(nil, errors.New("fail to get hash by number")).AnyTimes()
	bChain.EXPECT().GetBlockByHash(gomock.Any()).Return(nil, errors.New("fail to get block by hash")).AnyTimes()

	cBase := chainbase.NewMock(bChain, bCache)
	p.sync = New(cBase, p2pService)

	// Mock pob.
	go func() {
		for {
			select {
			case b := <-p.sync.ValidBlock():
				p.pobDoVerifyBlock(b)
			case <-p.quitCh:
				return
			}
		}
	}()
}

func TestBasic(t *testing.T) {
	// Ignore all ilog msgs.
	ilog.DefaultLogger().SetLevel(9)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mCenter := msgCenter{}
	dCenter := dataCenter{
		blocks:     make(map[int]*blockcache.BlockCacheNode),
		blockHashs: make(map[string]int),
	}

	// 0 -> 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7
	//                |         |--> 20006 -> ... -> 20010
	//                |--> 10004 -> ... -> 10009
	dCenter.addChain(7)
	dCenter.addChain(3, magicHeight+9)
	dCenter.addChain(5, 2*magicHeight+10)

	peers := make([]*peer, leastNeighborNumber+1)
	for i := 0; i <= leastNeighborNumber; i++ {
		p := newPeer(&mCenter, &dCenter, "Peer"+strconv.Itoa(i))
		p.addBlocks(0, 0)
		p.setLib(0)
		p.NewSync(ctrl)
		defer p.close()
		peers[i] = p
	}

	for i := 1; i <= leastNeighborNumber; i++ {
		peers[i].addBlocks(1, 7)
	}
	time.Sleep(5 * time.Second)
	if getID(peers[0].bcacheHead().Block) != 7 {
		t.Fatalf("Peer0's head should be 7")
	}

	for i := 1; i <= leastNeighborNumber; i++ {
		peers[i].addBlocks(10004, 10009)
	}
	time.Sleep(5 * time.Second)
	if getID(peers[0].bcacheHead().Block) != 10009 {
		t.Fatalf("Peer0's head should be 100009")
	}

	for i := 1; i <= leastNeighborNumber; i++ {
		peers[i].addBlocks(20006, 20010)
	}
	time.Sleep(5 * time.Second)
	if getID(peers[0].bcacheHead().Block) != 20010 {
		t.Fatalf("Peer0's head should be 20010")
	}
}
