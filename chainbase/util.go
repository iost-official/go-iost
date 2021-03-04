package chainbase

import (
	"context"
	"fmt"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	"github.com/iost-official/go-iost/v3/ilog"
	rpcpb "github.com/iost-official/go-iost/v3/rpc/pb"
	"google.golang.org/grpc"
)

// SPVFetchInitialBlockFromSeed get the most recent voting block older than the 'syncFrom' block
// if 'syncFrom' is 0, fetch the most recent voting block
func SPVFetchInitialBlockFromSeed(server string, syncFrom int64) (*block.Block, error) {
	rpcConn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := rpcpb.NewApiServiceClient(rpcConn)
	if syncFrom == 0 {
		value, err := client.GetChainInfo(context.Background(), &rpcpb.EmptyRequest{})
		if err != nil {
			return nil, err
		}
		syncFrom = value.LibBlock
	}
	syncFrom = syncFrom / common.VoteInterval * common.VoteInterval
	b, err := client.GetRawBlockByNumber(context.Background(), &rpcpb.GetBlockByNumberRequest{Number: syncFrom, Complete: true})
	if err != nil {
		return nil, err
	}
	blk := &block.Block{}
	blk.FromPb(b.Block)
	if err := blk.VerifySelf(); err != nil {
		return nil, fmt.Errorf("invalid block: %v", err)
	}
	ilog.Info("fetched seed block ", syncFrom, ",hash:", common.Base58Encode(blk.HeadHash()))
	return blk, nil
}
