package p2p

import (
	"fmt"
	"testing"
)

func TestNewNode(t *testing.T) {
	config := &Config{
		ListenAddr: "0.0.0.0:8088",
	}
	node, err := NewNetService(config)
	fmt.Println(string(node.host.ID()), node.host.ID().Pretty())
	fmt.Println(len(string(node.host.ID())), len(node.host.ID().Pretty()))
	fmt.Println(err)
}
