package pow

//
//import (
//	"encoding/binary"
//	"fmt"
//	"github.com/iost-official/prototype/core"
//	"math"
//	"time"
//)
//
//const (
//	SampleSize   = 6
//	TargetPeriod = 1 * time.Minute
//)
//
//func MineHead(bh core.BlockHead, dif uint64) core.BlockHead {
//	for nonce := uint64(0); nonce < math.MaxUint64; nonce++ {
//		bh.Info = makeInfo(dif, nonce)
//		ac := binary.BigEndian.Uint64(bh.Hash()[0:8])
//		if ac < dif {
//			break
//		}
//	}
//	return bh
//}
//
//func GetDifficulty(chain core.BlockChain) uint64 {
//	dif, _ := parseInfo(chain.Top().Head)
//
//	if chain.Length() < SampleSize {
//		return dif
//	}
//
//	var t int64
//	l := chain.Length()
//
//	for i := 1; i <= 3; i++ {
//		blk, _ := chain.Get(l - i)
//		t += blk.Head.Time
//	}
//
//	for i := 4; i <= 6; i++ {
//		blk, _ := chain.Get(l - i)
//		t -= blk.Head.Time
//	}
//
//	meanTime := t / 9
//
//	var rate float32
//
//	rate = float32(meanTime) / float32(TargetPeriod.Nanoseconds())
//
//	if rate > 4 {
//		rate = 4
//	} else if rate < 1/4 {
//		rate = 1 / 4
//	}
//
//	ans := float32(dif) * rate
//
//	return uint64(ans)
//}
//
//func IsLegalBlock(block *core.Block, up core.UTXOPool) error {
//	if len(block.Head.Info) < 8 {
//		return fmt.Errorf("syntax error")
//	}
//	ac := binary.BigEndian.Uint64(block.Head.Hash()[0:8])
//	dif := binary.BigEndian.Uint64(block.Head.Info[0:8])
//	if ac < dif {
//		return fmt.Errorf("unsolved nonce")
//	}
//
//	var txp core.TxPoolImpl
//	txp.Decode(block.Content)
//	txs, err := txp.GetSlice()
//	if err != nil {
//		return fmt.Errorf("syntax error")
//	}
//	for _, tx := range txs {
//		if err := isLegalTx(tx, up); err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//func isLegalTx(tx core.Tx, up core.UTXOPool) error {
//	for _, in := range tx.Inputs {
//		if _, err := up.Find(in.UTXOHash); err != nil {
//			return err
//		}
//	}
//
//	return nil // TODO : run script
//}
