package main

import (
	"flag"

	"fmt"

	"github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/network/discover"
)

func main() {
	mode := flag.String("mode", "public", "operation mode: private | public")
	flag.Parse()
	fmt.Println("[WARNING] Running in " + *mode + " mode. ")
	network.NetMode = *mode

	bootnodeStart()
}

func initNetConf() *network.NetConfig {
	conf := &network.NetConfig{}
	conf.LogPath = "/tmp"
	conf.NodeTablePath = "/tmp"
	conf.ListenAddr = "0.0.0.0"
	return conf
}

type ipInfo struct {
	IP           string `json:"ip"`
	RegisterTime int64  `json:"register_time"`
	ConnTime     int64  `json:"conn_time"`
}

func bootnodeStart() {
	node, err := discover.ParseNode("84a8ecbeeb6d3f676da1b261c35c7cd15ae17f32b659a6f5ce7be2d60f6c16f9@0.0.0.0:30304")
	if err != nil {
		fmt.Printf("parse boot node got err:%v\n", err)
	}
	conf := initNetConf()
	conf.NodeID = string(node.ID)
	baseNet, err := network.NewBaseNetwork(conf)
	if err != nil {
		fmt.Println("NewBaseNetwork ", err)
		return
	}
	ch, err := baseNet.Listen(node.TCP)
	if err != nil {
		fmt.Println("Init ", err)
		return
	}
	fmt.Println("server starting", node.Addr())
	for {
		select {
		case <-ch:
		}
	}
}
