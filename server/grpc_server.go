package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/besanh/chatbot_gpt/common/response"
	log "github.com/besanh/logger/logging/slog"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	Server *grpc.Server
	Port   string
}

func NewGRPCServer(port string) *GRPCServer {
	options := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
		),
	}
	grpcServer := grpc.NewServer(options...)

	// Reflection
	reflection.Register(grpcServer)

	return &GRPCServer{
		Server: grpcServer,
		Port:   port,
	}
}

func (s *GRPCServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.Port))
	if err != nil {
		return err
	}

	log.Infof("server runs on %s", s.Port)

	return s.Server.Serve(lis)
}

func (s *GRPCServer) HandleMetadata(ctx context.Context, r *http.Request) metadata.MD {
	md := make(map[string]string)

	return metadata.New(md)
}

func (s *GRPCServer) HandleError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
	code, response := response.HandleGRPCErrResponse(err)
	s.HTTPErrorHandler(ctx, marshaler, writer, request, code, response)
}

func (s *GRPCServer) HandleMatchHeaders(key string) (string, bool) {
	switch key {
	default:
		return key, false
	}
}

func (s *GRPCServer) HTTPErrorHandler(ctx context.Context, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, codes codes.Code, resp any) {
	const fallback = `{"code": 13, "message": "failed to marshal error message"}`

	buf, merr := marshaler.Marshal(resp)
	if merr != nil {
		grpclog.Infof("failed to marshal error message %q: %v", resp, merr)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			grpclog.Infof("failed to write response: %v", err)
		}
		return
	}

	w.Header().Del("Trailer")
	w.Header().Del("Transfer-Encoding")
	w.Header().Set("Content-Type", "application/json")
	st := runtime.HTTPStatusFromCode(codes)
	w.WriteHeader(st)
	if _, err := w.Write(buf); err != nil {
		grpclog.Infof("failed to write response: %v", err)
	}
}
