package main

import (
	"fmt"
	"BlockChainFramework/src/iosbase"
)

//import "BlockChainFramework/src/iosbase"


type A struct {
	a int
}
type B struct {
	b int
}

type AB struct {
	A
	B
}

func showA(xx A) {
	fmt.Println(xx.a)
}

func main() {
	fmt.Println("hello world")

	privkey := iosbase.Sha256([]byte{12,1})
	fmt.Println(iosbase.ToHex(privkey))

	sig := iosbase.Sign(iosbase.Sha256([]byte{3}), privkey)
	fmt.Println(iosbase.ToHex(sig))

	pubkey := iosbase.CalcPubkey(privkey)
	fmt.Println(iosbase.ToHex(pubkey))


	ab := AB{A{1}, B{2}}
	showA(ab.A)
}
