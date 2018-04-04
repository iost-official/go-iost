package p2p_test

import (
	"p2p"
	"testing"
)

func TestNetwork(t *testing.T) {
	nn := p2p.NewNaiveNetwork()
	nn.Send(p2p.Request{
		Time:    1,
		From:    "test1",
		To:      "test2",
		ReqType: 1,
		Body:    []byte{},
	})
}
