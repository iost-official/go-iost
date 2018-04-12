package main

import (
	"sync"

	"github.com/iost-official/prototype/console"

	_ "github.com/iost-official/prototype/common"
	_ "github.com/iost-official/prototype/core"
	_ "github.com/iost-official/prototype/event"
	_ "github.com/iost-official/prototype/iostdb"
	_ "github.com/iost-official/prototype/log"
	_ "github.com/iost-official/prototype/p2p"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		console.Listen()
		wg.Done()
	}()
	wg.Wait()
}
