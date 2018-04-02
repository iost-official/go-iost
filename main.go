package main

import (
	"fmt"

	_ "github.com/iost-official/PrototypeWorks/iosbase"
	"github.com/iost-official/PrototypeWorks/protocol"
)

type Node struct {
	protocol.ReplicaImpl
}

func main() {
	fmt.Println("hello world")

}
