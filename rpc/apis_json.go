package rpc

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/ilog"
)

// JSONServer json rpc server
type JSONServer struct {
	endPoint string
	jsonPort string
	srv      *http.Server
}

// NewJSONServer create json rpc server
func NewJSONServer(_global global.BaseVariable) *JSONServer {
	endPoint := strconv.Itoa(_global.Config().RPC.GRPCPort)
	if !strings.HasPrefix(endPoint, ":") {
		endPoint = "localhost:" + endPoint
	}

	jsonPort := strconv.Itoa(_global.Config().RPC.JSONPort)
	if !strings.HasPrefix(jsonPort, ":") {
		jsonPort = ":" + jsonPort
	}

	return &JSONServer{
		endPoint: endPoint,
		jsonPort: jsonPort,
	}
}

// Start start json rpc server
func (j *JSONServer) Start() error {
	go func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		mux := runtime.NewServeMux(runtime.WithMarshalerOption("*", &runtime.JSONPb{OrigName: true, EmitDefaults: true}))
		opts := []grpc.DialOption{grpc.WithInsecure()}

		err := RegisterApisHandlerFromEndpoint(ctx, mux, j.endPoint, opts)
		if err != nil {
			ilog.Errorf("NewJSONServer error: %v", err)
			return
		}

		j.srv = &http.Server{
			Addr:    j.jsonPort,
			Handler: mux,
		}

		j.srv.ListenAndServe()
	}()
	ilog.Info("JSON RPC server start")
	return nil
}

// Stop stop json rpc server
func (j *JSONServer) Stop() {
	err := j.srv.Shutdown(context.Background())
	if err != nil {
		ilog.Errorf("JSON RPC Stop error: %v", err)
	}
}
