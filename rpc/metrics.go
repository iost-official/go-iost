package rpc

import (
	"context"
	"strings"

	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/metrics"
	"google.golang.org/grpc"
)

var (
	requestCounter = metrics.NewCounter("iost_rpc_request", []string{"method"})
)

func metricsUnaryMiddleware(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	i := strings.LastIndex(info.FullMethod, "/")
	ilog.Debugf("receive rpc request: %s, request: %v", info.FullMethod[i+1:], req)
	requestCounter.Add(1, map[string]string{"method": info.FullMethod[i+1:]})
	return handler(ctx, req)
}

func metricsStreamMiddleware(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	i := strings.LastIndex(info.FullMethod, "/")
	ilog.Debugf("receive rpc stream: %s", info.FullMethod[i+1:])
	requestCounter.Add(1, map[string]string{"method": info.FullMethod[i+1:]})
	return handler(srv, ss)
}
