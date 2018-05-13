package network

import (
	"testing"
	"time"

	"fmt"

	"io/ioutil"
	"math/rand"
	"os"
	"strconv"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/network/discover"
	"github.com/iost-official/prototype/params"
	. "github.com/smartystreets/goconvey/convey"
)

//for boot test server

func initNetConf() *NetConifg {
	conf := &NetConifg{}
	conf.SetLogPath("iost_log_")
	tablePath, _ := ioutil.TempDir(os.TempDir(), "iost_node_table_"+strconv.Itoa(int(time.Now().UnixNano())))
	conf.SetNodeTablePath(tablePath)
	conf.SetListenAddr("127.0.0.1")
	return conf
}

//start boot node
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
		go router.Run()
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

		router.FilteredChan(Filter{AcceptType: []ReqType{ReqDownloadBlock}})
		router.FilteredChan(Filter{AcceptType: []ReqType{ReqBlockHeight}})
		go router.Run()
		rs = append(rs, router)
	}
	time.Sleep(15 * time.Second)

	return rs
}

func TestRouterImpl_Broadcast(t *testing.T) {
	routers := newRouters(3)
	height := uint64(32)
	deltaHeight := uint64(5)
	net0 := routers[0].(*RouterImpl).base.(*BaseNetwork)
	net1 := routers[1].(*RouterImpl).base.(*BaseNetwork)
	net2 := routers[2].(*RouterImpl).base.(*BaseNetwork)
	broadHeight := message.Message{Body: common.Uint64ToBytes(height), ReqType: int32(ReqBlockHeight), From: net2.localNode.String()}

	Convey("broadcast block", t, func() {
		//broadcast block height test
		go routers[2].Broadcast(broadHeight)
		time.Sleep(10 * time.Second)
		//check app msg chan
		select {
		case data := <-routers[1].(*RouterImpl).filterMap[1]:
			So(common.BytesToUint64(data.Body), ShouldEqual, height)
		}
		So(len(routers[1].(*RouterImpl).base.(*BaseNetwork).NodeHeightMap), ShouldEqual, 1)
		//download block request test

		net2.SetNodeHeightMap(net0.localNode.String(), height+uint64(rand.Int63n(int64(deltaHeight))))
		net2.SetNodeHeightMap(net1.localNode.String(), height+deltaHeight)
		go net2.Download(height, height+deltaHeight)
		for i := 0; i < (int(deltaHeight)+1)*MaxDownloadRetry; i++ {
			select {
			case data := <-routers[0].(*RouterImpl).filterMap[0]:
				So(common.BytesToUint64(data.Body), ShouldBeGreaterThan, height-1)
			case data := <-routers[1].(*RouterImpl).filterMap[0]:
				So(common.BytesToUint64(data.Body), ShouldBeGreaterThan, height-1)
			}
		}
		//	cancel download block test
	})

}
