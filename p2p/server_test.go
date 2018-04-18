package p2p

import (
	"testing"

	"strings"

	"time"

	"github.com/iost-official/prototype/iostdb"
	"github.com/iost-official/prototype/params"
)

func startServer(addr string) *Server {
	config := Config{ListenAddr: addr}
	srv := &Server{Config: config}
	srv.MaxPeers = 3
	srv.Name = addr
	srv.NodeDatabase = "p2p_test_" + addr
	srv.nodeTable, _ = iostdb.NewLDBDatabase(srv.NodeDatabase, 0, 0)
	for _, v := range params.TestnetBootnodes {
		strArr := strings.Split(v, "@")
		srv.nodeTable.Put([]byte(strArr[0]), []byte(strArr[1]))
	}
	return srv
}

func startBootServers() []*Server {
	servers := make([]*Server, 0)
	for _, addr := range params.TestnetBootnodes {
		strArr := strings.Split(addr, "@")
		server := startServer(strArr[1])
		server.Start()
		servers = append(servers, server)
	}
	return servers
}

func TestServer_Start(t *testing.T) {

	servers := startBootServers()

	config := Config{ListenAddr: "127.0.0.1:30403"}
	srv := &Server{Config: config}
	srv.MaxPeers = 3
	srv.Name = "test"
	srv.NodeDatabase = "p2p_test"
	srv.nodeTable, _ = iostdb.NewLDBDatabase(srv.NodeDatabase, 0, 0)
	for _, v := range params.TestnetBootnodes {
		strArr := strings.Split(v, "@")
		srv.nodeTable.Put([]byte(strArr[0]), []byte(strArr[1]))
	}
	srv.Start()
	time.Sleep(20 * time.Second)
	defer srv.Stop()
	defer func(servers []*Server) {
		for _, v := range servers {
			v.Stop()
		}
	}(servers)
}
