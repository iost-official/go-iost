package iosbase

import (
	"fmt"
	"testing"
)

func TestBinaryFromBase58(t *testing.T) {
	fmt.Println("============Test of Binary struct")
	bin := BinaryFromBase58("16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM")
	fmt.Printf("base58Decode > %v\n", ToHex(bin.Bytes()))
	fmt.Printf("")
}
