package network

import (
	"os"
	"testing"
	"time"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/network/discover"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBaseNetwork_AllNodesExcludeAddr(t *testing.T) {
	Convey("AllNodesExcludeAddr", t, func() {
		baseNet, _ := NewBaseNetwork(&NetConfig{RegisterAddr: registerAddr, ListenAddr: "127.0.0.1", NodeTablePath: "iost_db_"})
		iter := baseNet.nodeTable.NewIterator()
		for iter.Next() {
			baseNet.nodeTable.Delete(iter.Key())
		}
		iter.Release()
		So(iter.Error(), ShouldBeNil)

		baseNet.nodeTable.Put([]byte(registerAddr), common.IntToBytes(2))

		arr, err := baseNet.AllNodesExcludeAddr("")
		So(err, ShouldBeNil)
		So(len(arr), ShouldEqual, 1)
		So(arr[0], ShouldEqual, registerAddr)

		arr2, err := baseNet.AllNodesExcludeAddr(registerAddr)
		So(err, ShouldBeNil)
		So(len(arr2), ShouldEqual, 0)
	})
}

var registerAddr = "127.0.0.1:30304"

func cleanLDB() {
	os.RemoveAll("iost_db_")
	os.RemoveAll("iost_db_1")
	os.RemoveAll("iost_db_2")
	os.RemoveAll("iost_db_2")
	os.RemoveAll("iost_node_table_")
}

func TestBaseNetwork_recentSentLoop(t *testing.T) {
	Convey("recentSentLoop", t, func() {
		cleanLDB()
		baseNet, _ := NewBaseNetwork(&NetConfig{RegisterAddr: registerAddr, ListenAddr: "127.0.0.1", NodeTablePath: "iost_db_"})
		baseNet.RecentSent.Store("test_expired", time.Now().Add(-(MsgLiveThresholdSeconds+1)*time.Second))
		baseNet.RecentSent.Store("test_not_expired", time.Now())
		go func() {
			baseNet.recentSentLoop()
		}()
		time.Sleep(20 * time.Millisecond)
		_, ok1 := baseNet.RecentSent.Load("test_expired")
		_, ok2 := baseNet.RecentSent.Load("test_not_expired")
		So(ok1, ShouldBeFalse)
		So(ok2, ShouldBeTrue)
		cleanLDB()
	})
}

func TestBaseNetwork_isRecentSent(t *testing.T) {
	Convey("isRecentSent", t, func() {
		cleanLDB()
		baseNet, _ := NewBaseNetwork(&NetConfig{RegisterAddr: registerAddr, ListenAddr: "127.0.0.1", NodeTablePath: "iost_db_"})
		msg := message.Message{From: "sender", Time: time.Now().UnixNano(), To: "192.168.1.34:20003", Body: []byte{22, 11, 125}, TTL: 2}
		is := baseNet.isRecentSent(msg)
		So(is, ShouldBeFalse)
		is = baseNet.isRecentSent(msg)
		So(is, ShouldBeTrue)
		msg.TTL = msg.TTL - 1
		is = baseNet.isRecentSent(msg)
		So(is, ShouldBeTrue)
		msg.To = msg.To + msg.To
		is = baseNet.isRecentSent(msg)
		So(is, ShouldBeFalse)
		cleanLDB()
	})
}

var addresses = []string{
	"127.0.0.1:30301",
	"127.0.0.1:30302",
	"127.0.0.1:30303",
	"127.0.0.1:30305",
	"127.0.0.1:30306",
	"127.0.0.1:30307",
	"127.0.0.1:30308",
	"127.0.0.1:30309",
	"19.192.22.23:30310",
	"18.192.22.23:30311",
}

func TestBaseNetwork_findNeighbours(t *testing.T) {
	Convey("findNeighbours", t, func() {
		cleanLDB()
		bn, _ := NewBaseNetwork(&NetConfig{RegisterAddr: "127.0.0.1:30304", ListenAddr: "127.0.0.1", NodeTablePath: "iost_db_"})
		for _, addr := range addresses {
			bn.putNode(addr)
		}
		var neighbourLen int
		bn.neighbours.Range(func(k, v interface{}) bool {
			neighbourLen++
			return true
		})
		So(neighbourLen, ShouldEqual, discover.MaxNeighbourNum)
		_, ok1 := bn.neighbours.Load("@" + addresses[7])
		So(ok1, ShouldBeFalse)
		_, ok2 := bn.neighbours.Load("@" + addresses[6])
		So(ok2, ShouldBeFalse)

		bn.neighbours.Range(func(k, v interface{}) bool {
			node := v.(*discover.Node)
			bn.neighbours.Delete(node.String())
			bn.nodeTable.Delete([]byte(node.Addr()))
			return true
		})

		neighbourLen = 0
		bn.neighbours.Range(func(k, v interface{}) bool {
			neighbourLen++
			return true
		})
		So(neighbourLen, ShouldEqual, 0)
		cleanLDB()
	})

}

func TestBaseNetwork_putNode(t *testing.T) {
	Convey("putNode", t, func() {
		cleanLDB()
		bn, _ := NewBaseNetwork(&NetConfig{RegisterAddr: "127.0.0.1:30304", ListenAddr: "127.0.0.1", NodeTablePath: "iost_db_"})

		bn.putNode(addresses[0])
		b, err := bn.nodeTable.Get([]byte(addresses[0]))
		So(err, ShouldBeNil)
		So(common.BytesToInt(b), ShouldEqual, NodeLiveCycle)

		bn.putNode(addresses[0])
		b, err = bn.nodeTable.Get([]byte(addresses[0]))
		So(err, ShouldBeNil)
		So(common.BytesToInt(b), ShouldEqual, NodeLiveCycle)

		b, err = bn.nodeTable.Get([]byte(addresses[1]))
		So(err, ShouldNotBeNil)
		So(common.BytesToInt(b), ShouldEqual, 0)
		cleanLDB()
	})
}

func TestBaseNetwork_registerLoop(t *testing.T) {
	Convey("registerLoop", t, func() {
		cleanLDB()
		cleanLDB()
	})
}
