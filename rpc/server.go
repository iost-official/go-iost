package rpc

import (
	"fmt"
	"net"
	"strconv"

	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/core/new_blockcache"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
	//"github.com/iost-official/Go-IOS-Protocol/core/new_txpool"
	"strings"

	"github.com/iost-official/Go-IOS-Protocol/core/new_block"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"google.golang.org/grpc"
)

var bc blockcache.BlockCache
var txdb tx.TxDB

//var txpool txpool.TxPool
var bchain block.Chain
var visitor *database.Visitor

// Server gRPC entry
func Server(bcache blockcache.BlockCache, _global global.BaseVariable) error {
	txdb = _global.TxDB()
	//txpool=tp
	bchain = _global.BlockChain()
	bc = bcache
	visitor = database.NewVisitor(0, _global.StateDB())
	port := strconv.Itoa(_global.Config().RPC.Port)
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

	RegisterApisServer(s, newRPCServer())

	go s.Serve(lis)

	return nil
}
