package rpc

import (
	"context"
	"net"
	"net/http"

	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/rpc/pb"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

// Server is the rpc server including grpc server and json gateway server.
type Server struct {
	grpcAddr   string
	grpcServer *grpc.Server

	gatewayAddr   string
	gatewayServer *http.Server
}

func New() *Server {
	return &Server{}
}

func (s *Server) Start() error {
	if err := s.startGrpc(); err != nil {
		return err
	}
	return s.startGateway()
}

func (s *Server) startGrpc() error {
	lis, err := net.Listen("tcp", s.grpcAddr)
	if err != nil {
		return err
	}
	s.grpcServer = grpc.NewServer(grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(metricsMiddleware)))
	apiService := NewAPIService(nil, nil, nil, nil)
	rpcpb.RegisterApiServiceServer(s.grpcServer, apiService)
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			ilog.Fatalf("start grpc failed. err=%v", err)
		}
	}()
	return nil
}

func (s *Server) startGateway() error {
	mux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard,
		&runtime.JSONPb{OrigName: true, EmitDefaults: true}))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := rpcpb.RegisterApiServiceHandlerFromEndpoint(context.Background(), mux, s.grpcAddr, opts)
	if err != nil {
		return err
	}
	s.gatewayServer = &http.Server{
		Addr:    s.gatewayAddr,
		Handler: mux,
	}
	go func() {
		if err := s.gatewayServer.ListenAndServe(); err != nil {
			ilog.Fatalf("start gateway failed. err=%v", err)
		}
	}()
	return nil
}

func (s *Server) Stop() {
	s.gatewayServer.Shutdown(nil)
	s.grpcServer.GracefulStop()
}
