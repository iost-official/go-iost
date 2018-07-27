package main

import (
	. "github.com/iost-official/Go-IOS-Protocol/console"
	"sync"
)

func main() {
	var cli Console
	cli.Init(
		Help(),
		Connect(),
		Createblockchain(),
		Getbalance(),
		Printchain(),
		Send(),
		Exit(),
	)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		cli.Listen("> ")
		wg.Done()
	}()
	wg.Wait()
}
