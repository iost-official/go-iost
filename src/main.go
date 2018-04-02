package main

import (
	"fmt"
	"github.com/iost-official/prototype/p2p"
)

func main() {
	fmt.Println("hello world")
	network := p2p.NewNaiveNetwork()
	var myNet p2p.Network
	myNet = network
}
