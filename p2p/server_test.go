package p2p

import (
	"errors"
	"testing"

	"fmt"

	"strconv"
	"strings"

	"time"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/params"
	"github.com/magiconair/properties/assert"
	. "github.com/smartystreets/goconvey/convey"
)

//for boot test server
//优化测试参数 ping，同步node
//启动server 返回数据通道
//一键启动多个server，
var servers []*Server

func StartBootServers() error {

	for _, encodeAddr := range params.TestnetBootnodes {
		addr := extractAddrFromBoot(encodeAddr)
		if addr != "" {
			s, err := NewServer()
			if err != nil {
				return errors.New("new server encountered err " + fmt.Sprintf("%v", err))
			}
			servers = append(servers, s)
			arr := strings.Split(addr, ":")
			port, err := strconv.Atoi(arr[1])
			if err != nil {
				return errors.New("extract port encountered err " + fmt.Sprintf("%v", err))
			}

			recvCh, err := s.Listen(uint16(port))
			go func() {
				for v := range recvCh {
					fmt.Println("== %v", v)
				}
			}()

			//	todo got recv
		}
	}
	time.Sleep(100 * time.Second)
	return nil
}

func TestServer_Listen(t *testing.T) {
	err := StartBootServers()
	assert.Equal(t, nil, err)
}

func TestServer_allNodesExcludeAddr(t *testing.T) {
	Convey("", t, func() {
		s1, err := NewServer()
		s1.Listen(3003)
		So(err, ShouldEqual, nil)
		s1.nodeTable.Put([]byte("test node 1"), common.IntToBytes(0))
		s1.nodeTable.Put([]byte("test node 2"), common.IntToBytes(0))
		nodes, err := s1.allNodesExcludeAddr("test node 1")
		So(err, ShouldEqual, nil)
		So(string(nodes), ShouldContainSubstring, "test node 2")
		So(string(nodes), ShouldNotContainSubstring, "test node 1")
	})
}

func TestServer_rePickSeedAddr(t *testing.T) {
	Convey("rePick SeedAddr", t, func() {
		s1, err := NewServer()
		s1.Listen(3003)
		So(err, ShouldEqual, nil)
		s1.nodeTable.Put([]byte("test node 1"), common.IntToBytes(0))
		s1.rePickSeedAddr()
		So(s1.seedAddr, ShouldContainSubstring, "test node 1")
	})
}
