package rpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"strings"
)

func Server(port string) error {

	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	if s == nil {
		return fmt.Errorf("failed to rpc NewServer")
	}

	RegisterCliServer(s, newRpcServer())

	go s.Serve(lis)

	return nil
}
