package synchro

import (
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/iost-official/go-iost/consensus/synchro/pb"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

// Constant of blockhash sync
const (
	BlockHashLeastNeighborNumber = 2
	BlockHashExpiredSeconds      = 60
)

// BlockHash return the block hash with the Peers that have it.
type BlockHash struct {
	Hash   []byte
	Number int64
	PeerID []p2p.PeerID
}

type blockHashs struct {
	hashs map[int64][]byte
	time  int64
}

// blockHashSync is responsible for maintaining the recent blockhash status of neighbor nodes.
type blockHashSync struct {
	p                  p2p.Service
	newBlockHashCh     chan *BlockHash
	neighborBlockHashs map[p2p.PeerID]*blockHashs
	mutex              *sync.RWMutex

	msg1Ch chan p2p.IncomingMessage
	msg2Ch chan p2p.IncomingMessage

	quitCh chan struct{}
	done   *sync.WaitGroup
}

func newBlockHashSync(p p2p.Service) *blockHashSync {
	b := &blockHashSync{
		p:                  p,
		newBlockHashCh:     make(chan *BlockHash, 1024),
		neighborBlockHashs: make(map[p2p.PeerID]*blockHashs),
		mutex:              new(sync.RWMutex),

		msg1Ch: p.Register("new block hash", p2p.NewBlockHash),
		msg2Ch: p.Register("sync block hash response", p2p.SyncBlockHashResponse),

		quitCh: make(chan struct{}),
		done:   new(sync.WaitGroup),
	}

	b.done.Add(3)
	go b.newBlockHashController()
	go b.syncBlockHashResponseController()
	go b.expirationController()

	return b
}

func (b *blockHashSync) Close() {
	close(b.quitCh)
	b.done.Wait()
	ilog.Infof("Stopped block hash sync.")
}

// NewBlockHashs will return received new block hash.
func (b *blockHashSync) NewBlockHashs() <-chan *BlockHash {
	return b.newBlockHashCh
}

// NeighborBlockHashs will return all block hashs of neighbor nodes between start height and end height.
// Both start and end are included.
func (b *blockHashSync) NeighborBlockHashs(start, end int64) <-chan *BlockHash {
	ch := make(chan *BlockHash, 1024)
	go func() {
		for num := start; num <= end; num++ {
			hashs := make(map[string]*BlockHash)
			b.mutex.RLock()
			for peerID, blockHashs := range b.neighborBlockHashs {
				h, ok := blockHashs.hashs[num]
				if !ok {
					continue
				}
				key := string(h)
				if blockHash, ok := hashs[key]; ok {
					blockHash.PeerID = append(blockHash.PeerID, peerID)
				} else {
					hashs[key] = &BlockHash{
						Hash:   blockHashs.hashs[num],
						Number: num,
						PeerID: []p2p.PeerID{peerID},
					}
				}
			}
			b.mutex.RUnlock()

			for _, blockHash := range hashs {
				if len(blockHash.PeerID) >= BlockHashLeastNeighborNumber {
					ch <- blockHash
				}
			}
		}
		close(ch)
	}()
	return ch
}

func (b *blockHashSync) RequestBlockHash(start, end int64) {
	ilog.Debugf("Syncing block hash in [%v %v]...", start, end)

	blockHashQuery := &msgpb.BlockHashQuery{
		ReqType: msgpb.RequireType_GETBLOCKHASHES,
		Start:   start,
		End:     end,
		Nums:    nil,
	}
	msg, err := proto.Marshal(blockHashQuery)
	if err != nil {
		ilog.Errorf("Marshal sync block hash message failed: %v", err)
		return
	}
	b.p.Broadcast(msg, p2p.SyncBlockHashRequest, p2p.UrgentMessage)
}

func (b *blockHashSync) handleNewBlockHash(msg *p2p.IncomingMessage) {
	blockInfo := &msgpb.BlockInfo{}
	err := proto.Unmarshal(msg.Data(), blockInfo)
	if err != nil {
		ilog.Warnf("Unmarshal new block hash failed: %v", err)
		return
	}

	blockHash := &BlockHash{
		Hash:   blockInfo.Hash,
		Number: blockInfo.Number,
		PeerID: []p2p.PeerID{msg.From()},
	}
	b.newBlockHashCh <- blockHash
}

func (b *blockHashSync) newBlockHashController() {
	for {
		select {
		case msg := <-b.msg1Ch:
			b.handleNewBlockHash(&msg)
		case <-b.quitCh:
			b.done.Done()
			return
		}
	}
}

func (b *blockHashSync) handleSyncBlockHashResponse(msg *p2p.IncomingMessage) {
	if msg.Type() != p2p.SyncBlockHashResponse {
		ilog.Warnf("Expect the type %v, but get a unexpected type %v", p2p.SyncBlockHashResponse, msg.Type())
		return
	}

	blockHashResponse := &msgpb.BlockHashResponse{}
	err := proto.Unmarshal(msg.Data(), blockHashResponse)
	if err != nil {
		ilog.Warnf("Unmarshal block hash response failed: %v", err)
		return
	}

	// TODO: Prevent neighbor node attacks

	if len(blockHashResponse.BlockInfos) > maxSyncRange {
		ilog.Warnf("BlockInfos length %v exceed maxSyncRange %v", len(blockHashResponse.BlockInfos), maxSyncRange)
		return
	}

	hashs := make(map[int64][]byte)
	for _, blockInfo := range blockHashResponse.BlockInfos {
		hashs[blockInfo.Number] = blockInfo.Hash
	}

	ilog.Debugf("Received block hash for peer %v, len %v.", msg.From().Pretty(), len(blockHashResponse.BlockInfos))

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if bHashs, ok := b.neighborBlockHashs[msg.From()]; ok {
		for k, v := range hashs {
			bHashs.hashs[k] = v
		}
	} else {
		b.neighborBlockHashs[msg.From()] = &blockHashs{
			hashs: hashs,
			time:  time.Now().Unix(),
		}
	}
}

func (b *blockHashSync) syncBlockHashResponseController() {
	for {
		select {
		case msg := <-b.msg2Ch:
			b.handleSyncBlockHashResponse(&msg)
		case <-b.quitCh:
			b.done.Done()
			return
		}
	}
}

func (b *blockHashSync) doExpiration() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	now := time.Now().Unix()
	for k, v := range b.neighborBlockHashs {
		if v.time+BlockHashExpiredSeconds < now {
			delete(b.neighborBlockHashs, k)
		}
	}
}

func (b *blockHashSync) expirationController() {
	for {
		select {
		case <-time.After(2 * time.Second):
			b.doExpiration()
		case <-b.quitCh:
			b.done.Done()
			return
		}
	}
}
