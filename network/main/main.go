package main

import (
	"flag"
	"time"

	"fmt"

	"github.com/iost-official/prototype/core/message"
	. "github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/network/discover"
)

var listenPort = flag.String("p", "30302", "go run main.go or go run main.go -p 30305 -s 127.0.0.1:30302")
var serverAddr = flag.String("s", "", "Specify local port 30302-30304, or other ports that have already started.")
var id = flag.String("i", "", "server tag id")
var conf = initNetConf()

func initNetConf() *NetConifg {
	conf := &NetConifg{}
	conf.SetLogPath("/tmp")
	conf.SetNodeTablePath("/tmp")
	conf.SetListenAddr("127.0.0.1")
	return conf
}

func main() {
	//testBaseNetwork()
	node, err := discover.ParseNode("84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9@127.0.0.1:30304")
	if err != nil {
		fmt.Errorf("parse boot node got err:%v", err)
	}
	router, _ := RouterFactory("base")
	conf := initNetConf()
	conf.SetNodeID(string(node.ID))
	baseNet, err := NewBaseNetwork(conf)
	if err != nil {
		fmt.Println("NewBaseNetwork ", err)
		return
	}
	err = router.Init(baseNet, node.TCP)
	if err != nil {
		fmt.Println("Init ", err)
		return
	}
	go router.Run()
	fmt.Println("server starting")
	select {}

}

func testBaseNetwork() {
	rs := make([]Router, 0)
	for i := 0; i < 3; i++ {
		router, _ := RouterFactory("base")
		baseNet, _ := NewBaseNetwork(&NetConifg{})
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
