package pow
//
//import (
//	"encoding/binary"
//	"github.com/iost-official/prototype/core"
//	"time"
//	. "github.com/iost-official/prototype/p2p"
//)
//
//type Pow struct {
//	core.Member
//	Recorder
//	BlockCacheImpl
//	Router
//
//	ExitSignal chan bool
//
//	chTx    chan core.Request
//	chBlock chan core.Request
//}
//
//const (
//	Port = 23333
//)
//
//func (p *Pow) Init(bc core.BlockChain, sp core.UTXOPool, network core.Network) error {
//	var err error
//	p.BlockCacheImpl = NewBlockCache(bc, sp)
//	p.Recorder = NewRecorder()
//	p.Router, err = RouterFactory("base")
//	if err != nil {
//		return err
//	}
//	p.Router.Init(network, 23333)
//
//	p.chTx, err = p.Router.FilteredChan(Filter{
//		AcceptType: []ReqType{ReqPublishTx},
//	})
//
//	p.chBlock, err = p.Router.FilteredChan(Filter{
//		AcceptType: []ReqType{ReqNewBlock},
//	})
//
//	return err
//
//}
//
//func (p *Pow) Run() {
//	go p.Router.Run()
//}
//
//func (p *Pow) Stop() {
//	close(p.chBlock)
//	close(p.chTx)
//	p.ExitSignal <- true
//}
//func (p *Pow) PublishTx(tx core.Tx) error {
//	err := p.Recorder.Add(tx)
//	if err != nil {
//		return err
//	}
//	p.Send(core.Request{
//		ReqType: int(ReqPublishTx),
//		From:    p.ID,
//		To:      "ALL",
//		Body:    tx.Encode(),
//	})
//	return nil
//}
//func (p *Pow) CheckTx(tx core.Tx) (core.TxStatus, error) {
//	return core.POOL, nil
//}
////func (p *Pow) GetStatus() (core.BlockChain, core.UTXOPool, error) {
////	return p.bc, p.pool, nil
////}
////func (p *Pow) GetCachedStatus() (core.BlockChain, core.UTXOPool, error) {
////	return p.BlockCacheImpl.LongestChain(), p.BlockCacheImpl.LongestPool(), nil
////}
//
//func parseInfo(head core.BlockHead) (difficulty, nonce uint64) {
//	difficulty = binary.BigEndian.Uint64(head.Info[0:8])
//	nonce = binary.BigEndian.Uint64(head.Info[8:16])
//	return
//}
//
//func makeInfo(dif, nonce uint64) []byte {
//	b1 := make([]byte, 8)
//	binary.BigEndian.PutUint64(b1, dif)
//	b2 := make([]byte, 8)
//	binary.BigEndian.PutUint64(b2, nonce)
//
//	return append(b1, b2...)
//}
//
//func (p *Pow) txListenLoop() {
//	for {
//		req, ok := <-p.chTx
//		if !ok {
//			return
//		}
//		var tx core.Tx
//		tx.Decode(req.Body)
//		p.PublishTx(tx)
//	}
//}
//
////func (p *Pow) blockLoop() {
////	for {
////		req, ok := <-p.chBlock
////		if !ok {
////			return
////		}
////		var blk core.Block
////		blk.Decode(req.Body)
////		p.BlockCacheImpl.Add(&blk)
////	}
////}
//
//// dep
//func (p *Pow) mineLoop() {
//	for {
//		select {
//		case <-p.ExitSignal:
//			return
//		default:
//			chain := p.BlockCacheImpl.LongestChain()
//			dif := GetDifficulty(chain)
//			content := p.Recorder.Pop()
//			blk := core.Block{
//				Version: Version,
//				Head: core.BlockHead{
//					Version:   Version,
//					ParentHash: chain.Top().HeadHash(),
//					TreeHash:  content.Hash(),
//					Time:      time.Now().UnixNano(),
//				},
//				Content: content.Encode(),
//			}
//			blk.Head = MineHead(blk.Head, dif)
//			p.PublishBlock(&blk)
//		}
//	}
//}
//
//func (p *Pow) PublishBlock(block *core.Block) {
//	p.Send(core.Request{
//		ReqType: int(ReqNewBlock),
//		From:    p.ID,
//		To:      "ALL",
//		Body:    block.Encode(),
//	})
//}
