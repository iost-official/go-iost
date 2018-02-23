package iost

import (
	"bytes"
	"math/big"
)

type State interface {
	GetValue() *big.Int
	GetScript() string
	GetHash() []byte
	Bytes() []byte
}

// 比特币UTXO的实现
type UTXO struct {
	value  uint64
	script string
	hash   []byte
}

func (u *UTXO) Bytes() []byte {
	var b Binary
	b.putULong(u.value)
	scriptBytes := bytes.NewBufferString(u.script).Bytes()
	b.putUInt(uint32(len(scriptBytes)))
	b.putBytes(scriptBytes)
	return b.Bytes()
}

func (u *UTXO) GetValue() *big.Int {
	return big.NewInt(int64(u.value))
}

func (u *UTXO) GetScript() string {
	return u.script
}

func (u *UTXO) GetHash() []byte {
	if u.hash == nil || len(u.hash) > 32 {
		u.hash = Sha256(u.Bytes())
	}
	return u.hash
}

func NewUtxo(value uint64, script string) UTXO {
	return UTXO{
		value:  value,
		script: script,
		hash:   nil,
	}
}

// end of UTXO
