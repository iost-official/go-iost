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

var recv chan message.Message

func newBootRouters() []Router {
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
func newRouters(n int) []Router {
	newBootRouters()
	rs := make([]Router, 0)
	for i := 0; i < n; i++ {
		router, _ := RouterFactory("base")
		baseNet, _ := NewBaseNetwork(&NetConifg{ListenAddr: "127.0.0.1"})
		router.Init(baseNet, uint16(30600+i))
		router.Run()

		rs = append(rs, router)
	}
	time.Sleep(15 * time.Second)
	recv, _ = rs[1].FilteredChan(Filter{AcceptType: []ReqType{ReqBlockHeight}})
	return rs
}
func TestRouterImpl_Broadcast(t *testing.T) {

	Convey("broadcast block height test", t, func() {
		routers := newRouters(3)
		net2 := routers[2].(*RouterImpl).base.(*BaseNetwork)
		height := uint64(32)
		broadHeight := message.Message{Body: common.Uint64ToBytes(height), ReqType: int32(ReqBlockHeight), From: net2.localNode.String()}

		routers[2].Broadcast(broadHeight)
		time.Sleep(10 * time.Second)
		//check app msg chan
		select {
		case data := <-recv:
			fmt.Println("recv msg = %v", data)
		}

		So(len(routers[1].(*RouterImpl).base.(*BaseNetwork).NodeHeightMap), ShouldEqual, 1)
	})

	Convey("download block", t, func() {

	})

}
