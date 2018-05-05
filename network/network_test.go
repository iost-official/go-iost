package network

import (
	"fmt"
	"testing"

	"github.com/iost-official/prototype/core/message"
	"github.com/stretchr/testify/assert"
)

func TestNetwork(t *testing.T) {
	nn, err := NewNaiveNetwork(3)
	if err != nil {
		t.Errorf("NewNaiveNetwork encounter err %+v", err)
		return
	}
	lis1, err := nn.Listen(11037)
	if err != nil {
		fmt.Println(err)
	}
	lis2, err := nn.Listen(11038)
	if err != nil {
		fmt.Println(err)
	}

	req := message.Message{
		Time:    1,
		From:    "test1",
		To:      "test2",
		ReqType: 1,
		Body:    []byte{1, 1, 2},
	}
	if err := nn.Broadcast(req); err != nil {
		t.Log("send request encounter err: %+v\n", err)
	}

	message := <-lis1
	assert.Equal(t, message, req)

	message = <-lis2
	assert.Equal(t, message, req)
}
