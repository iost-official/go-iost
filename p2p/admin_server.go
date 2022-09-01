package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/iost-official/go-iost/v3/ilog"

	peer "github.com/libp2p/go-libp2p/core/peer"
)

type adminServer struct {
	srv *http.Server
	pm  *PeerManager
}

func newAdminServer(port string, pm *PeerManager) *adminServer {
	mux := http.NewServeMux()
	as := &adminServer{
		srv: &http.Server{
			Addr:    "127.0.0.1:" + port,
			Handler: mux,
		},
		pm: pm,
	}
	as.registerRoute(mux)
	return as
}

func (as *adminServer) Start() {
	go func() {
		if err := as.srv.ListenAndServe(); err != http.ErrServerClosed {
			ilog.Errorf("p2p admin server start failed. err=%v", err)
		}
	}()
}

func (as *adminServer) Stop() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second) // nolint
	as.srv.Shutdown(ctx)
}

func (as *adminServer) registerRoute(mux *http.ServeMux) {
	mux.HandleFunc("/ping", Ping)
	mux.HandleFunc("/stats", as.Stats)
	mux.HandleFunc("/closepeer", as.ClosePeer)
	mux.HandleFunc("/closerandpeer", as.CloseRandPeer)
	mux.HandleFunc("/putipblack", as.PutIPBlack)
	mux.HandleFunc("/putpidblack", as.PutPIDBlack)
}

// Ping returns a "pong" to client.
func Ping(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("pong"))
}

// Stats returns information of neighbors.
func (as *adminServer) Stats(rw http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(as.pm.NeighborStat(), "", "  ")
	if err != nil {
		rw.Write([]byte(fmt.Sprintf("marshal error. err=%v", err)))
		return
	}
	rw.Write(bytes)
}

// CloseRandPeer close a random peer connection.
func (as *adminServer) CloseRandPeer(rw http.ResponseWriter, r *http.Request) {
	peers := as.pm.GetAllNeighbors()
	as.pm.RemoveNeighbor(peers[rand.Intn(len(peers))].id)
	rw.Write([]byte("ok"))
}

// ClosePeer close the peer's connection.
func (as *adminServer) ClosePeer(rw http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if len(params["pid"]) == 0 {
		rw.Write([]byte("params error. pid is missed."))
		return
	}
	pid := params["pid"][0]
	peerID, err := peer.Decode(pid)
	if err != nil {
		rw.Write([]byte("invalid peer id"))
		return
	}
	as.pm.RemoveNeighbor(peerID)
	rw.Write([]byte("ok"))
}

// PutPIDBlack puts a ip to black list.
func (as *adminServer) PutPIDBlack(rw http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if len(params["pid"]) == 0 {
		rw.Write([]byte("params error. pid is missed."))
		return
	}
	pid := params["pid"][0]
	peerID, err := peer.Decode(pid)
	if err != nil {
		rw.Write([]byte("invalid peer id"))
		return
	}
	as.pm.PutPIDToBlack(peerID)
	rw.Write([]byte("ok"))
}

// PutIPBlack puts a ip to black list.
func (as *adminServer) PutIPBlack(rw http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	if len(params["ip"]) == 0 {
		rw.Write([]byte("params error. ip is missed."))
		return
	}
	ip := params["ip"][0]
	as.pm.PutIPToBlack(ip)
	rw.Write([]byte("ok"))
}
