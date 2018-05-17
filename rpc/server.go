package rpc

import (
	"net"
	"google.golang.org/grpc"
	"fmt"
)

const (
	port = ":30303"
)

func Server() error {

	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	if s == nil {
		return fmt.Errorf("failed to rpc NewServer")
	}

	RegisterCliServer(s, newHttpServer())

	go s.Serve(lis)

	return nil
}
