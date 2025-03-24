package main

import (
	"context"
	grpcapi "github.com/HSE-RDBMS-course-work/kvstore/internal/api/grpc"
	"github.com/HSE-RDBMS-course-work/kvstore/internal/core"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
)

//todo package for getters with no "Get"

type SlogWrapper struct {
	log *slog.Logger
}

func (sw *SlogWrapper) Log(ctx context.Context, level logging.Level, msg string, fields ...any) {
	sw.log.Log(ctx, slog.Level(level), msg, fields)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})

	logger := slog.New(handler)

	store := core.NewStore()

	api := grpcapi.NewServer(store)

	recoveryFunc := func(p any) (err error) {
		log.Printf("panic: %v", p)
		return status.Error(codes.Internal, "internal server error")
	}

	recoveryMW := recovery.UnaryServerInterceptor(
		recovery.WithRecoveryHandler(recoveryFunc),
	)

	loggingMW := logging.UnaryServerInterceptor(
		&SlogWrapper{
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

	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("cannot listen: %s", err)
	}

	go func() {
		if err = server.Serve(listener); err != nil {
			log.Fatalf("cannot serve listener: %s", err)
		}
	}()

	<-ctx.Done()

	server.GracefulStop()
}
