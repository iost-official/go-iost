package main

import (
	"fmt"
	"p2p"
)

func main() {
	network := p2p.NewNaiveNetwork()
	var myNet p2p.Network
	myNet = network
	req,_ := myNet.Listen(11037)
	myNet.Send(p2p.Request{
		Time:    1,
		From:    "test1",
		To:      "test2",
		ReqType: 1,
		Body:    []byte{1,2,3},
	})
	message := <-req
	fmt.Printf("%+v\n", message)
}
