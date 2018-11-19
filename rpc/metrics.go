package rpc

import (
	"github.com/iost-official/go-iost/metrics"
	"google.golang.org/grpc"
)

var (
	requestCounter = metrics.NewCounter("iost_rpc_request", []string{"method"})
)

func metricsMiddleware(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	requestCounter.Add(1, map[string]string{"method": info.FullMethod})
	return handler(srv, ss)
}
