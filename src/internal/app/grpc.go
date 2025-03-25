package app

import (
	"context"
	"fmt"
	grpcapi "github.com/HSE-RDBMS-course-work/kvstore/internal/api/grpc"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
	"net"
)

type GRPCServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type slogWrapper struct {
	log *slog.Logger
}

func (sw *slogWrapper) Log(ctx context.Context, level logging.Level, msg string, fields ...any) {
	sw.log.Log(ctx, slog.Level(level), msg, fields)
}

func StartGRPCServer(cfg *GRPCServerConfig, api *grpcapi.Server, logger *slog.Logger) *grpc.Server {
	recoveryFunc := func(p any) (err error) { //todo use slog with context
		log.Printf("panic: %v", p)
		return status.Error(codes.Internal, "internal server error")
	}

	recoveryMW := recovery.UnaryServerInterceptor(
		recovery.WithRecoveryHandler(recoveryFunc),
	)

	loggingMW := logging.UnaryServerInterceptor(
		&slogWrapper{
			log: logger,
		},
	)

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recoveryMW,
			loggingMW,
		),
	)

	api.RegisterTo(server)

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("cannot listen: %s", err)
	}

	go func() {
		if err = server.Serve(listener); err != nil {
			log.Fatalf("cannot serve listener: %s", err)
		}
	}()

	return server
}
