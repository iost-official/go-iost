package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
)

func init() {
	lw := ilog.NewFileWriter("./logs")
	logger := ilog.New()
	// logger := ilog.NewConsoleLogger()
	logger.AddWriter(lw)
	ilog.InitLogger(logger)
}

func main() {
	seed := flag.String("seed", "", "seed node")
	port := flag.Int("port", 0, "port number")

	flag.Parse()

	config := &common.P2PConfig{
		ChainID: 111,
	}
	if *seed == "" {
		*seed = "/ip4/47.75.42.25/tcp/3389/ipfs/12D3KooWScGKxdJwgtSU6ShUaiS4nuad73z7wUbUcHEUVQNzomhg"
	}
	config.SeedNodes = []string{*seed}
	if *port <= 0 {
		*port = randomPort()
	}
	if *port <= 0 {
		ilog.Fatalf("invalid tcp port")
	}
	config.ListenAddr = "0.0.0.0:" + strconv.Itoa(*port)

	ns, err := p2p.NewNetService(config)
	if err != nil {
		ilog.Fatalf("create p2pservice failed. err=%v", err)
	}
	ns.Start()
	ct := NewChatter(ns)
	ct.Start()

	ilog.Infof("start. id=%s, addrs=%s", ns.ID(), ns.LocalAddrs())

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT)
	s := <-c
	ilog.Infof("received quit signal: %s", s)
	ct.Stop()
	ns.Stop()
}
