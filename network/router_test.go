package network

import (
	"testing"
	"time"

	"fmt"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/network/discover"
	"github.com/iost-official/prototype/params"
	. "github.com/smartystreets/goconvey/convey"
)

var recvHeightChMap map[int]chan message.Message
var recvDownloadChMap map[int]chan message.Message

//start boot node
func newBootRouters() []Router {
	recvHeightChMap = make(map[int]chan message.Message, 0)
	recvDownloadChMap = make(map[int]chan message.Message, 0)
	rs := make([]Router, 0)
	for _, encodeAddr := range params.TestnetBootnodes {
		node, err := discover.ParseNode(encodeAddr)
		if err != nil {
			fmt.Errorf("parse boot node got err:%v", err)
		}
		router, _ := RouterFactory("base")
		conf := initNetConf()
		conf.SetNodeID(string(node.ID))
		baseNet, err := NewBaseNetwork(conf)
		if err != nil {
			fmt.Println("NewBaseNetwork ", err)
		}
		err = router.Init(baseNet, node.TCP)
		if err != nil {
			fmt.Println("Init ", err)
		}
		router.Run()
	}
	return rs
}

//create n nodes
func newRouters(n int) []Router {
	newBootRouters()
	rs := make([]Router, 0)
	for i := 0; i < n; i++ {
		router, _ := RouterFactory("base")
		baseNet, _ := NewBaseNetwork(&NetConifg{ListenAddr: "127.0.0.1"})
		router.Init(baseNet, uint16(30600+i))
		router.Run()

		rs = append(rs, router)
		recv, _ := router.FilteredChan(Filter{AcceptType: []ReqType{ReqBlockHeight}})
		down, _ := router.FilteredChan(Filter{AcceptType: []ReqType{ReqDownloadBlock}})
		recvHeightChMap[i] = recv
		recvDownloadChMap[i] = down
	}
	time.Sleep(15 * time.Second)

	return rs
}

func TestRouterImpl_Broadcast(t *testing.T) {
	routers := newRouters(3)
	height := uint64(32)
	net0 := routers[0].(*RouterImpl).base.(*BaseNetwork)
	net1 := routers[1].(*RouterImpl).base.(*BaseNetwork)
	net2 := routers[2].(*RouterImpl).base.(*BaseNetwork)
	broadHeight := message.Message{Body: common.Uint64ToBytes(height), ReqType: int32(ReqBlockHeight), From: net2.localNode.String()}

	Convey("broadcast block height test", t, func() {

		routers[2].Broadcast(broadHeight)
		time.Sleep(10 * time.Second)
		//check app msg chan
		select {
		case data := <-recvHeightChMap[1]:
			fmt.Printf("recv msg = %v\n", data)
		}

		So(len(routers[1].(*RouterImpl).base.(*BaseNetwork).NodeHeightMap), ShouldEqual, 1)
	})

	Convey("download block request test", t, func() {
		net2.SetNodeHeightMap(net0.localNode.String(), height+2)
		net2.SetNodeHeightMap(net1.localNode.String(), height+5)

		routers[2].Download(height, height+5)
		select {
		case data := <-recvDownloadChMap[0]:
			fmt.Printf("node 0 receive download request (<=height+2) %v\n", data)
		case data := <-recvDownloadChMap[1]:
			fmt.Printf("node 1 receive download request (<=height+5)%v\n", data)
		}

	})

	Convey("cancel download block test", t, func() {
		//todo
	})

}
