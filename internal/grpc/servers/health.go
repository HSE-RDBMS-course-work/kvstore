package servers

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	pb "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthServer struct {
	*health.Server
}

func NewHealthServer() *HealthServer {
	return &HealthServer{
		Server: health.NewServer(),
	}
}

func (s *HealthServer) RegisterTo(server *grpc.Server) {
	pb.RegisterHealthServer(server, s)
}
