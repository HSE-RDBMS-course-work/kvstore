package servers

import (
	"context"
	"errors"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kvstore/internal/sl"
	"log/slog"
	"net"
	"runtime/debug"
	"time"
)

type slogWrapper struct {
	log *slog.Logger
}

func (sw *slogWrapper) Log(ctx context.Context, level logging.Level, msg string, fields ...any) {
	sw.log.Log(ctx, slog.Level(level), msg, fields)
}

type Config struct {
	Address string
	Timeout time.Duration
}

type Server struct {
	listener net.Listener
	*grpc.Server
}

func New(logger *slog.Logger, conf Config) (*Server, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	recoveryFunc := func(p any) (err error) {
		logger.Error("panic while handling grpc request",
			sl.Panic(p),
			slog.String("trace", string(debug.Stack())),
		)
		return status.Error(codes.Internal, "internal servers error")
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
		grpc.ConnectionTimeout(conf.Timeout),
	)

	listener, err := net.Listen("tcp", conf.Address)
	if err != nil {
		return nil, fmt.Errorf("failed create listener: %w", err)
	}

	return &Server{
		listener: listener,
		Server:   server,
	}, nil
}

func (s *Server) Run() error {
	return s.Server.Serve(s.listener)
}
