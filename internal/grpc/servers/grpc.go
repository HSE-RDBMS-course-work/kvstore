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
	Address           string
	ConnectionTimeout time.Duration
}

type Server struct {
	logger   *slog.Logger
	listener net.Listener
	*grpc.Server
}

func New(logger *slog.Logger, conf Config) (*Server, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	recoveryLogger := logger.With(sl.Component("grpc.Recovery"))

	recoveryFunc := func(p any) (err error) {
		recoveryLogger.Error("panic while handling grpc request",
			sl.Panic(p),
			slog.String("trace", string(debug.Stack())),
		)
		return status.Error(codes.Internal, "internal servers error")
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			recovery.UnaryServerInterceptor(
				recovery.WithRecoveryHandler(recoveryFunc),
			),
			logging.UnaryServerInterceptor(
				&slogWrapper{log: logger},
			),
		),
		grpc.ConnectionTimeout(conf.ConnectionTimeout),
	)

	listener, err := net.Listen("tcp", conf.Address)
	if err != nil {
		return nil, fmt.Errorf("failed create listener: %w", err)
	}

	logger = logger.With(sl.Component("grpc.Servers"))

	logger.Debug("created successfully", sl.Conf(conf))

	return &Server{
		logger:   logger,
		listener: listener,
		Server:   server,
	}, nil
}

func (s *Server) Run() error {
	s.logger.Info("starting server", slog.String("address", s.listener.Addr().String()))
	return s.Server.Serve(s.listener)
}

func (s *Server) Shutdown() error {
	s.logger.Info("shutting down")
	s.Server.GracefulStop()
	s.logger.Info("shut down gracefully")
	return nil
}
