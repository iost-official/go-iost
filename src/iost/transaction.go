package iost

import (
	"fmt"
)

type Transaction interface {
	Verify(pool StatePool) (bool, error)
	Transact(pool StatePool)
	Bytes() []byte
}

// 以下是btc交易的实现，目前只生成1对1的脚本

type BtcTrade struct {
	signature []byte
	pubkey    []byte
	input     []State
	output    []State
}

// 验证自身是否正确的逻辑，此函数只判断脚本解析器返回TRUE则成功，可以在这个接口实现gas等逻辑
// 这里使用了btcScriptor，一个只实现了P2PKH脚本的脚本解析器
// 一个Ethereum-like的verifier，则可以将value做为gas来使用，一定程度上提供了可扩展性
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

// 提供了与bitcoin不同的序列化方法，前64字节签名，34字节公钥，4字节inputAmount之后是input，4字节outputAmount之后是output
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
