package rpc

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/iost-official/Go-IOS-Protocol/core/global"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
)

type JSONServer struct {
	endPoint string
	jsonPort string
	srv      *http.Server
}

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

func (j *JSONServer) Start() error {
	go func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		mux := runtime.NewServeMux()
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
	return nil
}

func (j *JSONServer) Stop() {
	err := j.srv.Shutdown(context.Background())
	if err != nil {
		ilog.Errorf("JSON RPC Stop error: %v", err)
	}
}
