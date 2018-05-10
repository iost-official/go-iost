package network

import (
	"testing"
	"time"

	"strings"

	"strconv"

	"github.com/iost-official/prototype/core/message"
	"github.com/iost-official/prototype/params"
)

func newBootRouters() []Router {
	rs := make([]Router, 0)
	for _, encodeAddr := range params.TestnetBootnodes {
		addr := extractAddrFromBoot(encodeAddr)
		addrFrag := strings.Split(addr, ":")
		router, _ := RouterFactory("base")
		baseNet, _ := NewBaseNetwork(conf)
		port, _ := strconv.Atoi(addrFrag[1])
		router.Init(baseNet, uint16(port))
		router.Run()
	}
	return rs
}
func newRouters(n int) []Router {
	rs := make([]Router, 0)
	for i := 0; i < n; i++ {
		router, _ := RouterFactory("base")
		baseNet, _ := NewBaseNetwork(&NetConifg{})
		router.Init(baseNet, uint16(30602+i))
		router.Run()
		rs = append(rs, router)
	}
	return rs
}

func TestRouterImpl_Broadcast(t *testing.T) {
	req := message.Message{From: "sender", Time: time.Now().Unix(), To: "receiver", Body: []byte{22, 11, 125}}
	newBootRouters()
	routers := newRouters(3)
	time.Sleep(30 * time.Second)
	routers[2].Broadcast(req)
	time.Sleep(10 * time.Second)

}
