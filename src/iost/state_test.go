package iost

import (
	"testing"
	"fmt"
)

func TestNewUtxo(t *testing.T) {
	fmt.Println("============以下是UTXO的测试")
	script, _ := BtcLockScript("P2PKH", "3B7LCJa6Rx6RD5bZu1FNH7MQbGFi")
	utxo := NewUtxo(12450, script)
	fmt.Printf("此utxo的值 > %v\n", utxo.value)
	fmt.Printf("此utxo的hash > %v\n", ToHex(utxo.GetHash()))
}
