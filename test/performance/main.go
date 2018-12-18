package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/iost-official/go-iost/test/performance/call"
	_ "github.com/iost-official/go-iost/test/performance/handles/gobang"
	_ "github.com/iost-official/go-iost/test/performance/handles/luckyBet"
	_ "github.com/iost-official/go-iost/test/performance/handles/transfer"
)

func init() {
	log.SetOutput(os.Stdout)
}

var (
	defaultAmount  = 9999999999
	defaultTPS     = 9999999999
	defaultServers = []string{"3.0.81.219:30002", "3.0.192.236:30002"}
	defaultJob     = "gobang"
)

func main() {
	amount := flag.Int("a", defaultAmount, "tx amount")
	tps := flag.Int("t", defaultTPS, "tps")
	prepare := flag.Int("p", 1, "create account and deploy contract")
	s := flag.String("s", "", "rpc servers") // format: ip1:port1,ip2:port2
	job := flag.String("j", defaultJob, "tx job")

	flag.Parse()

	servers := defaultServers
	if *s != "" {
		servers = strings.Split(*s, ",")
	}

	fmt.Printf("\nsend %d %s transactions to %v, tps: %v prepare: %v\n\n", *amount, *job, servers, *tps, *prepare)

	call.InitClients(servers)

	log.Println("Start test!")
	start := time.Now()
	call.Run(*job, *amount, *tps, *prepare, false)
	log.Println("done. timecost=", time.Since(start))
}
