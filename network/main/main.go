package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/network"
)

var listenPort = flag.String("p", "30302", "go run main.go or go run main.go -p 30305 -s 127.0.0.1:30302")
var serverAddr = flag.String("s", "", "Specify local port 30302-30304, or other ports that have already started.")
var id = flag.String("i", "", "server tag id")
var conf = initNetConf()

func initNetConf() *network.NetConifg {
	conf := &network.NetConifg{}
	conf.SetLogPath("/tmp")
	conf.SetNodeTablePath("/tmp")
	return conf
}
func main() {
	testBaseNetwork()
	//testServer()
}

func testServer() {
	flag.Parse()
	s, err := network.NewServer(conf)

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
	req := message.Message{
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
			s.BroadcastCh <- network.MsgCh{Data: data, Priority: 1}
			time.Sleep(5 * time.Second)
		}
	}()
	// receive message from other nodes
	for {
		select {
		case r := <-s.RecvCh:
			fmt.Printf("<<<<<<<<<: %s %v %v\n", string(r.Body), r.From, r.Time)
			nodes, _ := s.AllNodesExcludeAddr("")
			fmt.Printf("[nodetabls] =  %+v \n\n", nodes)
		}
	}
}

func testBaseNetwork() {
	rs := make([]network.Router, 0)
	for i := 0; i < 3; i++ {
		router, _ := network.RouterFactory("base")
		baseNet, _ := network.NewBaseNetwork(&network.NetConifg{})
		router.Init(baseNet, uint16(30302+i))
		router.Run()
		rs = append(rs, router)
	}
	go func() {
		req := message.Message{From: "sender", Time: time.Now().UnixNano(), To: "127.0.0.1:30303", Body: []byte{22, 11, 125}}
		for {
			//rs[1].Send(req)
			req.Body = append(req.Body, []byte{11}...)
			rs[2].Broadcast(req)
			time.Sleep(5 * time.Second)
		}
	}()
	for {
		select {}
	}

}
