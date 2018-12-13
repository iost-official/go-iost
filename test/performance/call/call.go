package call

import (
	"log"
	"sync"

	"github.com/iost-official/go-iost/ilog"
)

var handles = make(map[string]Handler)

// Run ...
func Run(handleType string, iterNum int, parallelNum int, address string, flag bool) {
	results := make(chan string)

	handle, exist := handles[handleType]
	if !exist {
		log.Println("There is not handle")
		return
	}

	err := handle.Init(address, parallelNum)
	if err != nil {
		log.Println("Init error")
		return
	}

	err = handle.Publish()
	if err != nil {
		log.Println("Publish error", err)
		return
	}

	var iWaitGroup sync.WaitGroup
	iWaitGroup.Add(iterNum)

	for i := 0; i < iterNum; i++ {
		go func(it int) {
			var waitGroup sync.WaitGroup

			for j := 0; j < parallelNum; j++ {
				ilog.Info("para: ", j)
				waitGroup.Add(1)
				go func(jj int) {
					Handle(handle, jj, results)
					waitGroup.Done()
				}(j)
			}
			waitGroup.Wait()
			iWaitGroup.Done()
		}(i)
	}

	go func() {
		iWaitGroup.Wait()
		close(results)
	}()

	Display(results, flag)
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
