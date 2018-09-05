package pob

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus/synchronizer"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/blockcache"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/core/txpool"
	"github.com/iost-official/Go-IOS-Protocol/db"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/metrics"
	"github.com/iost-official/Go-IOS-Protocol/p2p"
	"github.com/iost-official/Go-IOS-Protocol/vm"
)

var (
	metricsGeneratedBlockCount = metrics.NewCounter("iost_pob_generated_block", nil)
	metricsConfirmedLength     = metrics.NewGauge("iost_pob_confirmed_length", nil)
	metricsTxSize              = metrics.NewGauge("iost_block_tx_size", nil)
	metricsMode                = metrics.NewGauge("iost_node_mode", nil)
)

var (
	errSingle     = errors.New("single block")
	errDuplicate  = errors.New("duplicate block")
	errTxHash     = errors.New("wrong txs hash")
	errMerkleHash = errors.New("wrong tx receipt merkle hash")
)

var blockReqTimeout = 3 * time.Second

//PoB is a struct that handles the consensus logic.
type PoB struct {
	account         *account.Account
	baseVariable    global.BaseVariable
	blockChain      block.Chain
	blockCache      blockcache.BlockCache
	txPool          txpool.TxPool
	p2pService      p2p.Service
	synchronizer    synchronizer.Synchronizer
	verifyDB        db.MVCCDB
	produceDB       db.MVCCDB
	blockReqMap     *sync.Map
	exitSignal      chan struct{}
	chRecvBlock     chan p2p.IncomingMessage
	chRecvBlockHash chan p2p.IncomingMessage
	chQueryBlock    chan p2p.IncomingMessage
	chGenBlock      chan *block.Block
}

// NewPoB init a new PoB.
func NewPoB(account *account.Account, baseVariable global.BaseVariable, blockCache blockcache.BlockCache, txPool txpool.TxPool, p2pService p2p.Service, synchronizer synchronizer.Synchronizer, witnessList []string) *PoB {
	p := PoB{
		account:         account,
		baseVariable:    baseVariable,
		blockChain:      baseVariable.BlockChain(),
		blockCache:      blockCache,
		txPool:          txPool,
		p2pService:      p2pService,
		synchronizer:    synchronizer,
		verifyDB:        baseVariable.StateDB(),
		produceDB:       baseVariable.StateDB().Fork(),
		blockReqMap:     new(sync.Map),
		exitSignal:      make(chan struct{}),
		chRecvBlock:     p2pService.Register("consensus channel", p2p.NewBlock, p2p.SyncBlockResponse),
		chRecvBlockHash: p2pService.Register("consensus block head", p2p.NewBlockHash),
		chQueryBlock:    p2pService.Register("consensus query block", p2p.NewBlockRequest),
		chGenBlock:      make(chan *block.Block, 10),
	}
	staticProperty = newStaticProperty(p.account, witnessList)
	return &p
}

//Start make the PoB run.
func (p *PoB) Start() error {
	go p.messageLoop()
	go p.blockLoop()
	go p.scheduleLoop()
	return nil
}

//Stop make the PoB stop.
func (p *PoB) Stop() {
	close(p.exitSignal)
	close(p.chRecvBlock)
	close(p.chGenBlock)
}

func (p *PoB) messageLoop() {
	for {
		if p.baseVariable.Mode() != global.ModeInit {
			break
		}
		time.Sleep(time.Second)
	}
	for {
		select {
		case incomingMessage, ok := <-p.chRecvBlockHash:
			if !ok {
				ilog.Infof("chRecvBlockHash has closed")
				return
			}
			var blkHash message.BlockHash
			err := proto.Unmarshal(incomingMessage.Data(), &blkHash)
			if err != nil {
				continue
			}
			go p.handleRecvBlockHash(&blkHash, incomingMessage.From())
		case incomingMessage, ok := <-p.chQueryBlock:
			if !ok {
				ilog.Infof("chQueryBlock has closed")
				return
			}
			var rh message.RequestBlock
			err := rh.Decode(incomingMessage.Data())
			if err != nil {
				continue
			}
			go p.handleBlockQuery(&rh, incomingMessage.From())
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) handleRecvBlockHash(blkHash *message.BlockHash, peerID p2p.PeerID) {
	_, ok := p.blockReqMap.Load(string(blkHash.Hash))
	if ok {
		ilog.Info("block in block request map, block hash: ", blkHash.Hash)
		return
	}
	_, err := p.blockCache.Find(blkHash.Hash)
	if err == nil {
		ilog.Info("duplicate block")
		return
	}
	blkReq := &message.RequestBlock{
		BlockHash: blkHash.Hash,
	}
	bytes, err := proto.Marshal(blkReq)
	if err != nil {
		ilog.Debugf("fail to verify blocks, %v", err)
		return
	}
	p.blockReqMap.Store(string(blkHash.Hash), time.AfterFunc(blockReqTimeout, func() {
		p.blockReqMap.Delete(string(blkHash.Hash))
	}))
	p.p2pService.SendToPeer(peerID, bytes, p2p.NewBlockRequest, p2p.UrgentMessage)
	blkByte, err := proto.Marshal(blkHash)
	if err != nil {
		ilog.Error("fail to encode block hash")
		return
	}
	p.p2pService.Broadcast(blkByte, p2p.NewBlockHash, p2p.UrgentMessage)
}

func (p *PoB) handleBlockQuery(rh *message.RequestBlock, peerID p2p.PeerID) {
	var b []byte
	var err error
	node, err := p.blockCache.Find(rh.BlockHash)
	if err == nil {
		b, err = node.Block.Encode()
		if err != nil {
			ilog.Errorf("Fail to encode block: %v, err=%v", rh.BlockNumber, err)
			return
		}
		p.p2pService.SendToPeer(peerID, b, p2p.NewBlock, p2p.UrgentMessage)
		return
	}
	ilog.Infof("failed to get block from blockcache. err=%v", err)
	b, err = p.blockChain.GetBlockByteByHash(rh.BlockHash)
	if err != nil {
		ilog.Warnf("failed to get block from blockchain. err=%v", err)
		return
	}
	p.p2pService.SendToPeer(peerID, b, p2p.NewBlock, p2p.UrgentMessage)
}

func (p *PoB) handleGenesisBlock(blk *block.Block) error {
	if p.baseVariable.Mode() == global.ModeInit && p.baseVariable.BlockChain().Length() == 0 && common.Base58Encode(blk.HeadHash()) == p.baseVariable.Config().Genesis.GenesisHash {
		if !bytes.Equal(blk.CalculateTxsHash(), blk.Head.TxsHash) {
			return errTxHash
		}
		if !bytes.Equal(blk.CalculateMerkleHash(), blk.Head.MerkleHash) {
			return errMerkleHash
		}
		p.blockCache.AddGenesis(blk)
		err := p.blockChain.Push(blk)
		if err != nil {
			return fmt.Errorf("push block in blockChain failed, err: %v", err)
		}
		engine := vm.NewEngine(blk.Head, p.verifyDB)
		txr, err := engine.Exec(blk.Txs[0])
		if err != nil {
			return fmt.Errorf("exec tx failed, err: %v", err)
		}
		if !bytes.Equal(blk.Receipts[0].Encode(), txr.Encode()) {
			return fmt.Errorf("wrong tx receipt")
		}
		p.verifyDB.Tag(string(blk.HeadHash()))
		err = p.verifyDB.Flush(string(blk.HeadHash()))
		if err != nil {
			return fmt.Errorf("flush stateDB failed, err:%v", err)
		}
		err = p.baseVariable.TxDB().Push(blk.Txs, blk.Receipts)
		if err != nil {
			return fmt.Errorf("push tx and txr into TxDB failed, err:%v", err)
		}
		return nil
	}
	return fmt.Errorf("not genesis block")
}

func (p *PoB) blockLoop() {
	ilog.Infof("start blockloop")
	for {
		select {
		case incomingMessage, ok := <-p.chRecvBlock:
			if !ok {
				ilog.Infof("chRecvBlock has closed")
				return
			}
			var blk block.Block
			err := blk.Decode(incomingMessage.Data())
			if err != nil {
				ilog.Error("fail to decode block")
				continue
			}
			if incomingMessage.Type() == p2p.NewBlock {
				if p.baseVariable.Mode() == global.ModeInit {
					continue
				}
				ilog.Info("received new block, block number: ", blk.Head.Number)
				err = p.handleRecvBlock(&blk)
				timer, ok := p.blockReqMap.Load(string(blk.HeadHash()))
				if ok {
					timer.(*time.Timer).Stop()
					p.blockReqMap.Delete(string(blk.HeadHash()))
				} else {
					blkHash := &message.BlockHash{
						Height: blk.Head.Number,
						Hash:   blk.HeadHash(),
					}
					b, err := proto.Marshal(blkHash)
					if err != nil {
						ilog.Error("fail to encode block hash")
					} else {
						p.p2pService.Broadcast(b, p2p.NewBlockHash, p2p.UrgentMessage)
					}
				}
				if err != nil && err != errSingle {
					ilog.Debugf("received new block error, err:%v", err)
					continue
				}
				if err == errSingle {
					go p.synchronizer.CheckSync()
				}
			}
			if incomingMessage.Type() == p2p.SyncBlockResponse {
				ilog.Info("received sync block, block number: ", blk.Head.Number)
				if blk.Head.Number == 0 {
					err := p.handleGenesisBlock(&blk)
					if err != nil {
						ilog.Debugf("received genesis block error, err:%v", err)
					}
					continue
				} else {
					err = p.handleRecvBlock(&blk)
					if err != nil && err != errSingle && err != errDuplicate {
						ilog.Debugf("received sync block error, err:%v", err)
						continue
					}
					go p.synchronizer.OnBlockConfirmed(string(blk.HeadHash()), incomingMessage.From())
				}
			}
			go p.synchronizer.CheckSyncProcess()
		case blk, ok := <-p.chGenBlock:
			if !ok {
				ilog.Infof("chGenBlock has closed")
				return
			}
			ilog.Info("block from myself, block number: ", blk.Head.Number)
			err := p.handleRecvBlock(blk)
			if err != nil {
				ilog.Debugf("received new block error, err:%v", err)
				continue
			}
			go p.synchronizer.CheckGenBlock(blk.HeadHash())
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) scheduleLoop() {
	nextSchedule := timeUntilNextSchedule(time.Now().UnixNano())
	ilog.Infof("nextSchedule: %.2f", time.Duration(nextSchedule).Seconds())
	for {
		select {
		case <-time.After(time.Duration(nextSchedule)):
			ilog.Infof("nextSchedule: %.2f", time.Duration(nextSchedule).Seconds())
			ilog.Info(p.baseVariable.Mode())
			metricsMode.Set(float64(p.baseVariable.Mode()), nil)
			if witnessOfSec(time.Now().Unix()) == p.account.ID {
				if p.baseVariable.Mode() == global.ModeNormal {
					blk, err := generateBlock(p.account, p.blockCache.Head().Block, p.txPool, p.produceDB)
					ilog.Infof("gen block:%v", blk.Head.Number)
					if err != nil {
						ilog.Error(err.Error())
						continue
					}
					p.chGenBlock <- blk
					blkByte, err := blk.Encode()
					if err != nil {
						ilog.Error(err.Error())
						continue
					}
					go p.p2pService.Broadcast(blkByte, p2p.NewBlock, p2p.UrgentMessage)
				}
				time.Sleep(common.SlotLength * time.Second)
			}
			nextSchedule = timeUntilNextSchedule(time.Now().UnixNano())
			ilog.Infof("nextSchedule: %.2f", time.Duration(nextSchedule).Seconds())
		case <-p.exitSignal:
			return
		}
	}
}

func (p *PoB) handleRecvBlock(blk *block.Block) error {
	_, err := p.blockCache.Find(blk.HeadHash())
	if err == nil {
		return errDuplicate
	}
	err = verifyBasics(blk.Head, blk.Sign)
	if err != nil {
		return err
	}
	parent, err := p.blockCache.Find(blk.Head.ParentHash)
	p.blockCache.Add(blk)
	// staticProperty.addSlot(blk.Head.Time)
	if err == nil && parent.Type == blockcache.Linked {
		return p.addExistingBlock(blk, parent.Block)
	}
	return errSingle
}

func (p *PoB) addExistingBlock(blk *block.Block, parentBlock *block.Block) error {
	node, _ := p.blockCache.Find(blk.HeadHash())
	ok := p.verifyDB.Checkout(string(blk.HeadHash()))
	if !ok {
		p.verifyDB.Checkout(string(blk.Head.ParentHash))
		err := verifyBlock(blk, parentBlock, p.blockCache.LinkedRoot().Block, p.txPool, p.verifyDB)
		if err != nil {
			p.blockCache.Del(node)
			ilog.Error(err.Error())
			return err
		}
		p.verifyDB.Tag(string(blk.HeadHash()))
	}

	h := p.blockCache.Head()
	if node.Number > h.Number {
		p.txPool.AddLinkedNode(node, node)
	} else {
		p.txPool.AddLinkedNode(node, h)
	}

	p.blockCache.Link(node)
	p.updateInfo(node)
	for child := range node.Children {
		p.addExistingBlock(child.Block, node.Block)
	}
	return nil
}

func (p *PoB) updateInfo(node *blockcache.BlockCacheNode) {
	updateWaterMark(node)
	updateLib(node, p.blockCache)
}
