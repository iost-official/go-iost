package network

import (
	"testing"
	"time"

	"fmt"

	"os"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/network/discover"
	"github.com/iost-official/prototype/params"
	. "github.com/smartystreets/goconvey/convey"
)

func newBootRouters() []Router {
	rs := make([]Router, 0)
	for _, encodeAddr := range params.TestnetBootnodes {
		node, err := discover.ParseNode(encodeAddr)
		if err != nil {
			fmt.Errorf("parse boot node got err:%v", err)
		}
		router, _ := RouterFactory("base")
		conf.SetNodeID(string(node.ID))
		baseNet, _ := NewBaseNetwork(conf)
		router.Init(baseNet, node.TCP)
		router.Run()
	}
	return rs
}
func newRouters(n int) []Router {
	newBootRouters()
	fmt.Println("--------")
	os.Exit(0)
	rs := make([]Router, 0)
	for i := 0; i < n; i++ {
		router, _ := RouterFactory("base")
		baseNet, _ := NewBaseNetwork(&NetConifg{})
		router.Init(baseNet, uint16(30602+i))
		router.Run()
		rs = append(rs, router)
	}
	time.Sleep(30 * time.Second)
	return rs
}

func TestRouterImpl_Broadcast(t *testing.T) {
	Convey("broadcast test", t, func() {
		height := uint64(32)
		broadHeight := message.Message{Body: common.Uint64ToBytes(height), ReqType: int32(ReqBlockHeight)}
		newBootRouters()
		routers := newRouters(3)
		routers[2].Broadcast(broadHeight)
		//net2 := routers[2].(*RouterImpl).base.(*BaseNetwork)
		net1 := routers[1].(*RouterImpl).base.(*BaseNetwork)
		fmt.Println(fmt.Sprintf("%v", net1.NodeHeightMap))
		recv, err := routers[2].FilteredChan(Filter{AcceptType: []ReqType{ReqBlockHeight}})
		So(err, ShouldBeNil)
		select {
		case msg := <-recv:
			fmt.Println("%v", msg)
		case <-time.After(10 * time.Second):
			break
		}
	})
}
