package servers

import (
	"errors"
	"fmt"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"kvstore/internal/grpc/servers/interceptors"
	"kvstore/internal/sl"
	"log/slog"
	"net"
	"time"
)

type Config struct {
	Address           string
	Username          string
	Password          string
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
	if conf.Address == "" {
		return nil, errors.New("address required")
	}
	if conf.Username == "" {
		return nil, errors.New("username required")
	}
	if conf.Password == "" {
		return nil, errors.New("password required")
	}
	if conf.ConnectionTimeout <= 0 {
		conf.ConnectionTimeout = 0
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.NewRecovery(logger),
			interceptors.NewLogging(logger),
			interceptors.NewAuth(conf.Username, conf.Password, []string{
				healthpb.Health_Check_FullMethodName,
			}),
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
