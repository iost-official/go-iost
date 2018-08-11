package rpc
import (
	"context"
	"fmt"
	"reflect"

	"github.com/iost-official/Go-IOS-Protocol/account"
	"github.com/iost-official/Go-IOS-Protocol/common"
	"github.com/iost-official/Go-IOS-Protocol/consensus"
	"github.com/iost-official/Go-IOS-Protocol/core/block"
	"github.com/iost-official/Go-IOS-Protocol/core/message"
	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/core/tx"
	"github.com/iost-official/Go-IOS-Protocol/core/txpool"
	"github.com/iost-official/Go-IOS-Protocol/network"
	"github.com/iost-official/Go-IOS-Protocol/vm"
	"github.com/iost-official/Go-IOS-Protocol/vm/lua"
)

//go:generate mockgen -destination mock_rpc/mock_rpc.go -package rpc_mock github.com/iost-official/Go-IOS-Protocol/rpc CliServer
type MVCCDB interface {
	Get(table string, key string) (string, error)
}
var mvccdb MVCCDB
type RpcServer struct {
}

// newRpcServer
func newRpcServer() *RpcServer {
	s := &RpcServer{}
	return s
}

func (s *RpcServer) GetHeight(ctx context.Context,void *VoidReq) (*HeightRes,error){
	return &HeightRes{
		Height:block.BChain.Length(),
	}
}

func GetTxByHash(ctx context.Context,hash *HashReq) (*TxRaw, error) {
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	txHash := hash.Hash

	txDb := tx.TxDib
	if txDb == nil {
		panic(fmt.Errorf("TxDb should be nil"))
	}
	trx, err := txDb.(*tx.TxPoolDb).Get(txHash)
	if err != nil {
		return nil, err
	}
	txRaw:=trx.toTxRaw()
	return txRaw,nil
}

func GetBlockByHash(ctx context.Context,hash *HashReq) (*BlockRaw, error){
	if hash == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}

	bchain := block.BChain
	if bchain == nil {
		panic(fmt.Errorf("block.BChain cannot be nil"))
	}
	blk:=bchain.GetBlockByHash([]byte(hash.Hash))
	blkRaw:=blk.ToBlkRaw()
	return blkRaw,nil
}

func GetBlockByNum(ctx context.Context,num *NumReq) (*BlockRaw, error) {
	if num == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	num:=num.num
	bchain := block.BChain
	if bchain == nil {
		panic(fmt.Errorf("block.BChain cannot be nil"))
	}
	blk:=bchain.GetBlockByNumber(num)
	if blk==nil{
		return nil,nil
	}
	blkRaw:=blk.ToBlkRaw()
	return blkRaw,nil

}
func GetBalance(ctx context.Context,key *GetBalanceReq) (*GetBalanceRes, error) {
	if key == nil {
		return nil, fmt.Errorf("argument cannot be nil pointer")
	}
	pub := key.pubkey
	balance,err:=mvccdb.Get("","i-"+pub)
	if err!=nil{
		balance=err
	}
	return &GetBalanceRes{
		balance:balance,
	},nil
}
func GetState(ctx context.Context,key *GetStateReq) (*GetStateRes, error) {

}
func SendRawTx(ctx context.Context,rawTx *RawTxReq) (*SendRawTxRes, error){
	res := SendRawTxRes{}
	if rawTx == nil {
		return &ret, fmt.Errorf("argument cannot be nil pointer")
	}
	var trx tx.Tx
	err := trx.Decode(rawTx.data)
	if err != nil {
		return &ret, err
	}

	err = trx.VerifySelf() //verify Publisher and Signers
	if err != nil {
		return &ret, err
	}

	// add servi
	tx.RecordTx(trx, tx.Data.Self())

	//broadcast the tx
	router := network.Route
	if router == nil {
		panic(fmt.Errorf("network.Router shouldn't be nil"))
	}
	broadTx := message.Message{
		Body:    rawTx.data,//trx.Encode(),
		ReqType: int32(network.ReqPublishTx),
	}
	router.Broadcast(broadTx)
	Cons := consensus.Cons
	if Cons == nil {
		panic(fmt.Errorf("Consensus is nil"))
	}
	txpool.TxPoolS.AddTransaction(&broadTx)
	ret.hash = trx.Hash()
	return &ret, nil

}
func EstimateGas(ctx context.Context,rawTx *RawTxReq) (*GasRes, error){

}
