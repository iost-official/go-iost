package main

import (
	"github.com/iost-official/go-iost/test/performance/call"
	_ "github.com/iost-official/go-iost/test/performance/handles/transfer"
	"log"
	"os"
	"time"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	var iterNum = 800
	var parallelNum = 10
	var address = "localhost:30002"

	log.Println("Start test!")
	start := time.Now()
	call.Run("transfer", iterNum, parallelNum, address, false)

	log.Println("done. timecost=", time.Since(start))
}
