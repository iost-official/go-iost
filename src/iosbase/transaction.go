package iosbase

import (
	"fmt"
)

type Transaction interface {
	Verify(pool StatePool) (bool, error)
	Transact(pool StatePool)
	Bytes() []byte
}

// btc like transaction

type BtcTrade struct {
	signature []byte
	pubkey    []byte
	input     []State
	output    []State
}

func (bt *BtcTrade) Verify(pool StatePool) (isValid bool, err error) {
	for _, state := range bt.input {
		if pool.Get(state.GetHash()) != nil {
			inputScript := fmt.Sprint(Base58Encode(bt.signature), Base58Encode(bt.pubkey), state.GetScript())
			scriptor := NewBtcScriptor()
			result, err := scriptor.Run(inputScript)
			if err == nil && result == "TRUE" {
				continue
			} else {
				isValid = false
				return
			}
		} else {
			isValid = false
			err = fmt.Errorf("input not exist")
			return
		}
	}
	isValid = true
	err = nil
	return
}

// 将此transaction应用在statePool上
func (bt *BtcTrade) Transact(pool StatePool) {
	for _, state := range bt.input {
		pool.Del(state.GetHash())
	}
	for _, state := range bt.output {
		pool.Add(state)
	}
}

func (bt *BtcTrade) Bytes() []byte {
	var bin Binary
	bin.putBytes(bt.signature)
	bin.putBytes(bt.pubkey)
	bin.putUInt(uint32(len(bt.input)))
	for _, i := range bt.input {
		bin.put(i)
	}
	bin.putUInt(uint32(len(bt.output)))
	for _, i := range bt.output {
		bin.put(i)
	}
	return bin.Bytes()
}

// bitcoin transaction
