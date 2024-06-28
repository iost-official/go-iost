package main

import (
	"context"
	"fmt"

	"github.com/iost-official/go-iost/v3/common"
	"github.com/iost-official/go-iost/v3/core/block"
	rpcpb "github.com/iost-official/go-iost/v3/rpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func testVerifier() error {
	verifier := Verifier{}
	server := "54.180.196.80:30002"
	rpcConn, err := grpc.NewClient(server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	client := rpcpb.NewApiServiceClient(rpcConn)
	startBlock := (int64)(102504000)
	blk, err := client.GetRawBlockByNumber(context.Background(), &rpcpb.GetBlockByNumberRequest{Number: startBlock, Complete: true})
	if err != nil {
		return err
	}
	b := &block.Block{}
	b.FromPb(blk.Block)
	err = verifier.init(b)
	if err != nil {
		return err
	}
	// try to verify some blocks within next 1 hour
	for delta := 60; delta <= 7200; delta += 60 {
		blockNum := startBlock + int64(delta)
		blk, err := client.GetRawBlockByNumber(context.Background(), &rpcpb.GetBlockByNumberRequest{Number: blockNum, Complete: true})
		if err != nil {
			return err
		}
		b := &block.Block{}
		b.FromPb(blk.Block)
		blockList, err := client.GetBlockHeaderByRange(context.Background(), &rpcpb.GetBlockHeaderByRangeRequest{
			Start: blockNum + 1,
			End:   blockNum + 1 + 108,
		})
		if err != nil {
			return err
		}
		blkList := make([]*block.Block, 0, 108)
		for _, item := range blockList.BlockList {
			b := &block.Block{}
			b.FromPb(item)
			blkList = append(blkList, b)
		}
		if b.Head.Number%common.VoteInterval == 0 {
			err := verifier.updateEpoch(b, blkList)
			if err != nil {
				return err
			}
		} else {
			err := verifier.checkBlock(b, blkList)
			if err != nil {
				return err
			}
		}
	}
	fmt.Println("test done")
	return nil
}

func main() {
	err := testVerifier()
	if err != nil {
		panic(err)
	}
}
