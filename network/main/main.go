package main

import (
	"flag"
	"time"

	"fmt"

	"net"

	"strconv"

	"github.com/iost-official/prototype/core/message"
	. "github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/network/discover"
)

var serverAddr = flag.String("s", "", "server port 30304, or other ports that have already started.")
var conf = initNetConf()

func initNetConf() *NetConifg {
	conf := &NetConifg{}
	conf.SetLogPath("/tmp")
	conf.SetNodeTablePath("/tmp")
	conf.SetListenAddr("0.0.0.0")
	return conf
}

func main() {
	//bootnodeStart()
	testBaseNetwork()
	//testBootNodeConn()
}

func testBootNodeConn() {
	node, err := discover.ParseNode("84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9@18.219.254.124:30304")

	if err != nil {
		fmt.Errorf("parse boot node got err:%v", err)
	}
	conn, err := net.Dial("tcp", node.Addr())
	if err != nil {
		fmt.Println(err)
		return
	}
	conn.Write([]byte("111"))
	conn.Close()
	fmt.Println("conn to remote success!")
}

func bootnodeStart() {
	node, err := discover.ParseNode("84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9@0.0.0.0:30304")
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
	fmt.Println("server starting", node.Addr())
	select {}
}

func testBaseNetwork() {
	rs := make([]Router, 0)
	for i := 0; i < 2; i++ {
		router, _ := RouterFactory("base")
		baseNet, _ := NewBaseNetwork(&NetConifg{NodeTablePath: "node_table_" + strconv.Itoa(i), ListenAddr: "0.0.0.0"})
		err := router.Init(baseNet, uint16(20002+i))
		if err != nil {
			fmt.Println("init net got err", err)
			return
		}
		router.Run()
		rs = append(rs, router)
	}
	time.Sleep(15 * time.Second)
	go func() {
		req := message.Message{From: "sender", Time: time.Now().UnixNano(), To: "0.0.0.0:20003", Body: []byte{22, 11, 125}}
		for {
			//rs[1].Send(req)
			req.Body = append(req.Body, []byte{11}...)
			rs[0].Broadcast(req)
			time.Sleep(5 * time.Second)
		}
	}()
	for {
		select {}
	}

}
