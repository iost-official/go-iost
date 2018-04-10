package pow

import (
	"github.com/iost-official/prototype/core"
	"time"
	"encoding/binary"
	"fmt"
)

const (
	SampleSize   = 6
	TargetPeriod = 1 * time.Minute
)

func MineHead(bh core.BlockHead, dif uint64) core.BlockHead{
	for nonce := uint64(0); nonce < 0xFFFFFFFFFFFFFFFF; nonce ++ {
		bh.Info = makeInfo(dif, nonce)
		ac := binary.BigEndian.Uint64(bh.Hash()[0:8])
		if ac < dif {
			break
		}
	}
	return bh
}

func GetDifficulty(chain core.BlockChain) uint64 {
	dif, _ := parseInfo(chain.Top().Head)

	if chain.Length() < SampleSize {
		return dif
	}

	var t int64
	l := chain.Length()

	for i := 1; i <= 3; i ++ {
		blk, _ := chain.Get(l - i)
		t += blk.Head.Time
	}

	for i := 4; i <= 6; i ++ {
		blk, _ := chain.Get(l - i)
		t -= blk.Head.Time
	}

	meanTime := t / 9

	var rate float32

	rate = float32(meanTime) / float32(TargetPeriod.Nanoseconds())

	if rate > 4  {
		rate = 4
	} else if rate < 1/4 {
		rate = 1/4
	}

	ans := float32(dif) * rate

	return uint64(ans)
}

func IsLegalBlock(block core.Block) error {
	if len (block.Head.Info) < 8 {
		return fmt.Errorf("syntax error")
	}
	ac := binary.BigEndian.Uint64(block.Head.Hash()[0:8])
	dif := binary.BigEndian.Uint64(block.Head.Info[0:8])
	if ac > dif {
		return nil
	} else {
		return fmt.Errorf("unsolved nonce")
	}
}

func isLegalTx(tx core.Tx) error {
	return nil // TODO : complete it
}