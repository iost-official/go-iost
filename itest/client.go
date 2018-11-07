package itest

import (
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/rpc"
	"google.golang.org/grpc"
)

type Client struct {
	grpc rpc.ApisClient
	name string
	addr string
}

func (c *Client) GetGRPC() (rpc.ApisClient, error) {
	if c.grpc == nil {
		conn, err := grpc.Dial(c.addr)
		if err != nil {
			return nil, err
		}
		c.grpc = rpc.NewApisClient(conn)
		ilog.Infof("Create grpc connection with %v successful", c.addr)
	}
	return c.grpc, nil
}
