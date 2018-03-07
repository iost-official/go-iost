package iosbase

import (
	"testing"
	"fmt"
)

func TestNewUtxo(t *testing.T) {
	fmt.Println("============test of utxo")
	script, _ := BtcLockScript("P2PKH", "3B7LCJa6Rx6RD5bZu1FNH7MQbGFi")
	utxo := NewUtxo(12450, script)
	fmt.Printf("value of utxo > %v\n", utxo.value)
	fmt.Printf("hash > %v\n", ToHex(utxo.GetHash()))
}
