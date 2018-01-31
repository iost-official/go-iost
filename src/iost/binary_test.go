package iost

import (
	"testing"
	"fmt"
)

func TestBinaryFromBase58(t *testing.T) {
	fmt.Println("============以下是二进制的测试")
	bin := BinaryFromBase58("16UwLL9Risc3QfPqBUvKofHmBQ7wMtjvM")
	fmt.Printf("base58解码 > %v\n", ToHex(bin.Bytes()))
	fmt.Printf("")
}
