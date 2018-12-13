package call

import (
	"context"
	"fmt"
	"log"
	"sync"

	"golang.org/x/time/rate"
)

const (
	goroutineLimit = 2000
)

type semaphore chan struct{}

func (s semaphore) Acquire() {
	s <- struct{}{}
}

func (s semaphore) Release() {
	<-s
}

var handles = make(map[string]Handler)

// Run ...
func Run(handleType string, amount int, tps int, prepare int, flag bool) {
	// results := make(chan string)

	handle, exist := handles[handleType]
	if !exist {
		log.Println("There is not handle")
		return
	}

	if prepare > 0 {
		err := handle.Prepare()
		if err != nil {
			log.Println("prepare error: ", err)
			return
		}
	}
	limiter := rate.NewLimiter(rate.Limit(tps), 1)

	wg := new(sync.WaitGroup)
	wg.Add(amount)

	sem := make(semaphore, goroutineLimit)
	for i := 0; i < amount; i++ {
		err := limiter.Wait(context.Background())
		if err != nil {
			panic(err)
		}
		sem.Acquire()
		go func(it int) {
			defer sem.Release()
			defer wg.Done()
			// Handle(handle, it, results)
			_, err := handle.Run(it)
			if err != nil {
				fmt.Println(err)
			}
		}(i)
		if i%10000 == 0 {
			fmt.Printf("sent %d txs\n", i)
		}
	}

	wg.Wait()

	// Display(results, flag)
}

// Register ...
func Register(handleType string, handle Handler) {
	if _, exist := handles[handleType]; exist {
		log.Println("handle already registered")
		return
	}
	log.Println("Register HandleType:", handleType)
	handles[handleType] = handle
}
