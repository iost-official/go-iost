package main

import (
	"log"
	"os"
	"time"

	"github.com/iost-official/go-iost/test/performance/call"
	_ "github.com/iost-official/go-iost/test/performance/handles/gobang"
	_ "github.com/iost-official/go-iost/test/performance/handles/transfer"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	var iterNum = 1
	var parallelNum = 1
	var address = "localhost:30002"

	log.Println("Start test!")
	start := time.Now()
	call.Run("gobang", iterNum, parallelNum, address, false)

	log.Println("done. timecost=", time.Since(start))
}
