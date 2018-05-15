package rpc

//go:generate protoc --go_out=plugins=grpc:. ./cli.proto

//go:generate mockgen -destination mocks/mock_cli_client.go -package mock_rpc github.com/iost-official/prototype/rpc CliClient
//go:generate mockgen -destination mocks/mock_cli_server.go -package mock_rpc github.com/iost-official/prototype/rpc CliServer
