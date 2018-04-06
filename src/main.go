package main

import (
	"fmt"
	"p2p"
)

func main() {
	network := p2p.NewNaiveNetwork()
	var myNet p2p.Network
	myNet = network
	req1, _ := myNet.Listen(11037)
	req2, _ := myNet.Listen(11038)

	myNet.Send(p2p.Request{
		Time:    1,
		From:    "test1",
		To:      "test2",
		ReqType: 1,
		Body:    []byte{1, 2, 3},
	})

	message := <-req1
	fmt.Printf("%+v\n", message)
	message = <-req2
	fmt.Printf("%+v\n", message)
}
