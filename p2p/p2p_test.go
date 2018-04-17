package p2p

import (
	"fmt"
	"testing"
	"github.com/magiconair/properties/assert"
)

func TestNetwork(t *testing.T) {
	nn := NewNaiveNetwork()
	lis1, err := nn.Listen(11037)
	if err != nil {
		fmt.Println(err)
	}
	lis2, err := nn.Listen(11038)
	if err != nil {
		fmt.Println(err)
	}
	req := Request{
		Time:    1,
		From:    "test1",
		To:      "test2",
		ReqType: 1,
		Body:    []byte{1, 1, 2},
	}
	if err := nn.Send(req); err != nil {
		t.Log("send request encounter err: %v\n", err)
	}

	message := <-lis1
	assert.Equal(t, message, req)

	message = <-lis2
	assert.Equal(t, message, req)
}
