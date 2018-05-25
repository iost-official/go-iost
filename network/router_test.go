package network

import (
	"testing"
	"time"

	"fmt"

	"io/ioutil"
	"os"
	"strconv"

	"sync"

	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/network/discover"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRouterImpl_Init(t *testing.T) {
	//broadcast(t)
	router, _ := RouterFactory("base")
	baseNet, _ := NewBaseNetwork(&NetConifg{ListenAddr: "0.0.0.0"})
	router.Init(baseNet, 30601)
	Convey("init", t, func() {
		So(router.(*RouterImpl).port, ShouldEqual, 30601)
	})
}

func TestGetInstance(t *testing.T) {
	Convey("", t, func() {

		router, err := GetInstance(&NetConifg{NodeTablePath: "tale_test", ListenAddr: "127.0.0.1"}, "base", 30304)

		So(err, ShouldBeNil)
		So(router.(*RouterImpl).port, ShouldEqual, uint16(30304))
		So(Route.(*RouterImpl).port, ShouldEqual, uint16(30304))
	})
}

func initNetConf() *NetConifg {
	conf := &NetConifg{}
	conf.SetLogPath("iost_log_")
	tablePath, _ := ioutil.TempDir(os.TempDir(), "iost_node_table_"+strconv.Itoa(int(time.Now().UnixNano())))
	conf.SetNodeTablePath(tablePath)
	conf.SetListenAddr("0.0.0.0")
	return conf
}

//start boot node
func newBootRouters() []Router {
	rs := make([]Router, 0)
	node, err := discover.ParseNode("@127.0.0.1:30304")
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
	return rs
}

//create n nodes
func newRouters(n int) []Router {
	rs := make([]Router, 0)
	for i := 0; i < n; i++ {
		router, _ := RouterFactory("base")
		baseNet, _ := NewBaseNetwork(&NetConifg{RegisterAddr: "127.0.0.1:30304", ListenAddr: "127.0.0.1", NodeTablePath: "iost_db_" + strconv.Itoa(i)})
		router.Init(baseNet, uint16(20900+i))

		router.FilteredChan(Filter{AcceptType: []ReqType{ReqDownloadBlock}})
		router.FilteredChan(Filter{AcceptType: []ReqType{ReqBlockHeight}})
		go router.Run()
		rs = append(rs, router)
	}
	return rs
}

var m = message.Message{
	Time:    time.Now().Unix(),
	From:    "from",
	ReqType: int32(ReqBlockHeight),
	Body:    []byte("&network.NetConifg{LogPath:       logPath, NodeTablePath: nodeTablePath, NodeID:        nodeID, RegisterAddr:  regAddr, ListenAddr:    listenAddr},&network.NetConifg{LogPath:       logPath, NodeTablePath: nodeTablePath, NodeID:        nodeID, RegisterAddr:  regAddr, ListenAddr:    listenAddr},"),
}

//3 ms
func TestRouterImpl_Send(t *testing.T) {
	rs := newRouters(2)
	net0 := rs[0].(*RouterImpl).base.(*BaseNetwork)
	net1 := rs[1].(*RouterImpl).base.(*BaseNetwork)
	m.To = net0.localNode.String()
	net1.putNode(m.To)
	begin := time.Now().UnixNano()
	rs[1].Send(m)
	ch, _ := rs[0].FilteredChan(Filter{AcceptType: []ReqType{ReqBlockHeight}})
	wg := sync.WaitGroup{}
	wg.Add(1)
	for {
		select {
		case <-ch:
			wg.Done()
			goto Finish
		}
	}
Finish:
	wg.Wait()
	fmt.Println((time.Now().UnixNano()-begin)/int64(time.Millisecond), " ms/ per send")
	for _, r := range rs {
		r.Stop()
	}
}

//14ms
func TestRouterImpl_Broadcast(t *testing.T) {
	rs := newRouters(3)
	for k, route := range rs {
		for k2, route2 := range rs {
			if k != k2 {
				route.(*RouterImpl).base.(*BaseNetwork).putNode(route2.(*RouterImpl).base.(*BaseNetwork).localNode.Addr())
			}
		}
	}
	begin := time.Now().UnixNano()
	rs[0].Broadcast(m)
	ch1, _ := rs[1].FilteredChan(Filter{AcceptType: []ReqType{ReqBlockHeight}})
	ch2, _ := rs[2].FilteredChan(Filter{AcceptType: []ReqType{ReqBlockHeight}})

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		for {
			select {
			case <-ch1:
				wg.Done()
			case <-ch2:
				wg.Done()
			}
		}
	}()

	wg.Wait()
	fmt.Println((time.Now().UnixNano()-begin)/int64(time.Millisecond), " ms/ per send")
	for _, r := range rs {
		r.Stop()
	}
}
