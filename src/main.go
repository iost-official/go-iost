package main

import (
	"fmt"
	"p2p"
)

func main() {
	fmt.Println("hello world")
	network := p2p.NewNaiveNetwork()
	var myNet p2p.Network
	myNet = network
	myNet.Close(1)
}
