package main

import (
	"flag"
	"fmt"
	"time"

	"strconv"

	"github.com/iost-official/prototype/core"
	"github.com/iost-official/prototype/network"
)

var listenPort = flag.String("p", "30302", "go run main.go or go run main.go -p 30305 -s 127.0.0.1:30302")
var serverAddr = flag.String("s", "", "Specify local port 30302-30304, or other ports that have already started.")
var id = flag.String("i", "", "server tag id")

func main() {
	flag.Parse()
	s, err := network.NewServer()

	if err != nil {
		panic("node start panic:" + err.Error())
	}
	s.RemoteAddr = (*serverAddr)
	port, _ := strconv.Atoi(*listenPort)
	go func() {
		if _, err := s.Listen(uint16(port)); err != nil {
			panic(err)
		}
	}()
	time.Sleep(5 * time.Second)
	fmt.Println("server start....")
	req := core.Request{
		Time:    time.Now().Unix(),
		From:    "from:" + s.ListenAddr,
		To:      "to:" + s.RemoteAddr,
		ReqType: 1,
		Body:    []byte("transaction-" + (*id)),
	}
	data, _ := req.Marshal(nil)
	go func() {
		for {
			fmt.Println(">>>>>>>>>>>", string(data))
			s.BroadcastCh <- data
			time.Sleep(5 * time.Second)
		}
	}()
	// receive message from other nodes
	for {
		select {
		case r := <-s.RecvCh:
			fmt.Printf("<<<<<<<<<: %s %v\n", string(r.Body), r.From)
			nodes, _ := s.AllNodes()
			fmt.Printf("[nodetabls] =  %+v \n\n", nodes)
		}
	}

}
