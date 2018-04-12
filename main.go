package main

import (
	"sync"
	"github.com/iost-official/prototype/console"
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
