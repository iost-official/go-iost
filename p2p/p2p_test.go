package p2p

import (
	"fmt"
	"p2p"
	"testing"
)

func TestNetwork(t *testing.T) {
	nn := p2p.NewNaiveNetwork()
	lis1, err := nn.Listen(11037)
	if err != nil {
		fmt.Println(err)
	}
	lis2, err := nn.Listen(11038)
	if err != nil {
		fmt.Println(err)
	}
	nn.Send(p2p.Request{
		Time:    1,
		From:    "test1",
		To:      "test2",
		ReqType: 1,
		Body:    []byte{1, 1, 2},
	})

	message := <-lis1
	fmt.Printf("%+v\n", message)

	message = <-lis2
	fmt.Printf("%+v\n", message)
}
