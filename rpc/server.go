package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/iost-official/go-iost/core/blockcache"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/core/txpool"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/p2p"
	"github.com/iost-official/go-iost/rpc/pb"
	"github.com/rs/cors"
	"golang.org/x/net/netutil"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

const (
	maxConcurrentStreams = 200
	connectionLimit      = 128
)

// Server is the rpc server including grpc server and json gateway server.
type Server struct {
	grpcAddr   string
	grpcServer *grpc.Server

	gatewayAddr   string
	gatewayServer *http.Server
	allowOrigins  []string
}

// New returns a new rpc server instance.
func New(tp txpool.TxPool, bc blockcache.BlockCache, bv global.BaseVariable, p2pService p2p.Service) *Server {
	s := &Server{
		grpcAddr:     bv.Config().RPC.GRPCAddr,
		gatewayAddr:  bv.Config().RPC.GatewayAddr,
		allowOrigins: bv.Config().RPC.AllowOrigins,
	}
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(metricsMiddleware)),
		grpc.MaxConcurrentStreams(maxConcurrentStreams))
	apiService := NewAPIService(tp, bc, bv, p2pService)
	rpcpb.RegisterApiServiceServer(s.grpcServer, apiService)
	return s
}

// Start starts the rpc server.
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
	lis = netutil.LimitListener(lis, connectionLimit)
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			ilog.Fatalf("start grpc failed. err=%v", err)
		}
	}()
	return nil
}

func (s *Server) startGateway() error {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
		runtime.WithProtoErrorHandler(errorHandler))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := rpcpb.RegisterApiServiceHandlerFromEndpoint(context.Background(), mux, s.grpcAddr, opts)
	if err != nil {
		return err
	}
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "HEAD", "POST", "PUT", "DELETE"},
		AllowedOrigins: s.allowOrigins,
	})
	s.gatewayServer = &http.Server{
		Addr:    s.gatewayAddr,
		Handler: c.Handler(mux),
	}
	go func() {
		if err := s.gatewayServer.ListenAndServe(); err != http.ErrServerClosed {
			ilog.Fatalf("start gateway failed. err=%v", err)
		}
	}()
	return nil
}

func errorHandler(_ context.Context, _ *runtime.ServeMux, _ runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	w.WriteHeader(400)
	w.Write([]byte(fmt.Sprint(err)))
}

// Stop stops the rpc server.
func (s *Server) Stop() {
	s.gatewayServer.Shutdown(nil)
	s.grpcServer.GracefulStop()
}
