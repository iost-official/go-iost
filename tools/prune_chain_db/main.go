package main

import (
	"fmt"
	"os"

	"github.com/iost-official/go-iost/core/block"
)

func main() {
	from := os.Args[1]
	to := os.Args[2]
	chainDB, err := block.NewBlockChain(from)
	if err != nil {
		fmt.Println("cannot load chain", err)
		return
	}
	fmt.Println("start trim from", from, "to", to)
	err = chainDB.(*block.BlockChain).CopyLastNBlockTo(to, 10000)
	if err != nil {
		fmt.Println("cannot write chain", err)
		return
	}
	fmt.Println("trim chain done")
}
