package call

import (
	rpcpb "github.com/iost-official/go-iost/rpc/pb"
	"google.golang.org/grpc"
)

var clients []*GrpcClient

// GrpcClient wraps ApiServiceClient with a semaphore.
type GrpcClient struct {
	rpcpb.ApiServiceClient
	addr string
}

// Addr returns the address.
func (c *GrpcClient) Addr() string {
	return c.addr
}

// InitClients inits rpc clients.
func InitClients(addrs []string) {
	clients = make([]*GrpcClient, len(addrs))
	for i, addr := range addrs {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
		clients[i] = &GrpcClient{
			rpcpb.NewApiServiceClient(conn),
			addr,
		}
	}
}

// GetClient returns the corresponding grpc client.
func GetClient(i int) *GrpcClient {
	c := clients[i%len(clients)]
	return c
}
