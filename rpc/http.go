package rpc

import "context"

type HttpServer struct {
}

func (s *HttpServer) PublishTx(ctx context.Context, tx *Transaction) (*Response, error) {

	return nil, nil
}

func (s *HttpServer) GetContract(ctx context.Context, tx *ContractKey) (*Contract, error) {

	return nil, nil
}

func (s *HttpServer) GetBalance(ctx context.Context, tx *Key) (*Value, error) {

	return nil, nil
}

func (s *HttpServer) GetState(ctx context.Context, tx *Key) (*Value, error) {

	return nil, nil
}

func (s *HttpServer) GetBlock(ctx context.Context, tx *BlockKey) (*BlockInfo, error) {

	return nil, nil
}
