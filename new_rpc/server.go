package rpc

import (
	"fmt"
	"net"

	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/new_vm/database"
	//"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/new_tx"
	"google.golang.org/grpc"
 	"strings"
)

var bc blockcache.BlockCache
var txdb tx.TxDB

//var txpool txpool.TxPool
var bchain block.Chain
var visitor *database.Visitor
//func Server(port string, tp txpool.TxPool,bcache blockcache.BlockCache, _global global.Global) error {
func Server(port string, bcache blockcache.BlockCache, _global global.BaseVariable) error {
	txdb = _global.TxDB()
	//txpool=tp
	bchain = _global.BlockChain()
	bc = bcache
	visitor=database.NewVisitor(0,_global.StateDB())
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	lis, err := net.Listen("tcp4", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	if s == nil {
		return fmt.Errorf("failed to rpc NewServer")
	}

	RegisterApisServer(s, newRpcServer())

	go s.Serve(lis)

	return nil
}
