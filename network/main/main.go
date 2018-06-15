package main

import (
	"encoding/json"
	"flag"
	"net/http"

	"fmt"

	"github.com/iost-official/prototype/network"
	"github.com/iost-official/prototype/network/discover"
)

func main() {
	mode := flag.String("mode", "public", "operation mode: private | public | committee")
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

func dumpNodeServer(baseNet *network.BaseNetwork) {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	http.HandleFunc("/nodes", func(w http.ResponseWriter, r *http.Request) {
		ips, err := baseNet.AllNodesExcludeAddr("")
		resp := make(map[string]interface{})
		if err != nil {
			fmt.Println("get all nodes failed. err=", err)
			resp = map[string]interface{}{
				"status":  10001,
				"message": fmt.Sprintf("error=%v", err),
			}
		} else {
			resp = map[string]interface{}{
				"status":  0,
				"message": "success",
				"ips":     ips,
			}
		}

		b, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("json marshal error:", err)
		}
		w.Write(b)
	})
	go http.ListenAndServe(":30306", nil)

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
	dumpNodeServer(baseNet)
	for {
		select {
		case <-ch:
		}
	}

}
