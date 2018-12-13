package rpc

import (
	"context"
	"strings"

	"github.com/iost-official/go-iost/metrics"
	"google.golang.org/grpc"
)

var (
	requestCounter = metrics.NewCounter("iost_rpc_request", []string{"method"})
)

func metricsMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	i := strings.LastIndex(info.FullMethod, "/")
	//ilog.Debugf("receive rpc request: %s", info.FullMethod[i+1:])
	requestCounter.Add(1, map[string]string{"method": info.FullMethod[i+1:]})
	return handler(ctx, req)
}
