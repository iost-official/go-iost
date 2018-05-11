package network

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"io/ioutil"
	"os"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/params"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

//for boot test server
var servers []*Server

func initNetConf() *NetConifg {
	conf := &NetConifg{}
	conf.SetLogPath("iost_log_")
	tablePath, _ := ioutil.TempDir(os.TempDir(), "iost_node_table_"+strconv.Itoa(int(time.Now().UnixNano())))
	conf.SetNodeTablePath(tablePath)
	conf.SetListenAddr("127.0.0.1")
	return conf
}

func StartBootBaseNetWorks() error {
	for _, encodeAddr := range params.TestnetBootnodes {
		addr := extractAddrFromBoot(encodeAddr)
		if addr != "" {
			s, err := NewServer(initNetConf())
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
					fmt.Println("recv msg = ", v)
				}
			}()

			//	todo got recv
		}
	}
	time.Sleep(100 * time.Second)
	return nil
}

func TestBaseNetWork_Listen(t *testing.T) {
	err := StartBootBaseNetWorks()
	assert.Equal(t, nil, err)
}

func TestBaseNetWork_allNodesExcludeAddr(t *testing.T) {
	Convey("", t, func() {
		s1, err := NewServer(initNetConf())
		s1.Listen(3003)
		So(err, ShouldEqual, nil)
		s1.nodeTable.Put([]byte("test node 1"), common.IntToBytes(0))
		s1.nodeTable.Put([]byte("test node 2"), common.IntToBytes(0))
		nodes, err := s1.AllNodesExcludeAddr("test node 1")
		So(err, ShouldEqual, nil)
		So(string(nodes), ShouldContainSubstring, "test node 2")
		So(string(nodes), ShouldNotContainSubstring, "test node 1")
	})
}

func TestBaseNetWork_rePickSeedAddr(t *testing.T) {
	Convey("rePick SeedAddr", t, func() {
		s1, err := NewServer(initNetConf())
		s1.Listen(3003)
		So(err, ShouldEqual, nil)
		s1.nodeTable.Put([]byte("test node 1"), common.IntToBytes(0))
		s1.rePickSeedAddr()
		So(s1.seedAddr, ShouldContainSubstring, "test node 1")
	})
}
