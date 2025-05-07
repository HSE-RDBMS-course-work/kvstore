package servers

import (
	"context"
	"errors"
	pb "github.com/HSE-RDBMS-course-work/kvstore-proto/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kvstore/internal/raft"
)

type cluster interface {
	AcceptJoin(ctx context.Context, in raft.JoinToClusterIn) error
}

type RaftServer struct {
	pb.UnimplementedRaftServer
	cluster cluster
}

func NewRaftServer(cluster cluster) (*RaftServer, error) {
	if cluster == nil {
		return nil, errors.New("cluster is required")
	}

	return &RaftServer{
		cluster: cluster,
	}, nil
}

func (s *RaftServer) RegisterTo(server *grpc.Server) {
	pb.RegisterRaftServer(server, s)
}

func (s *RaftServer) Join(ctx context.Context, in *pb.JoinIn) (*pb.JoinOut, error) {
	err := s.cluster.AcceptJoin(ctx, raft.JoinToClusterIn{
		JoinerID:      raft.ServerID(in.JoinerId),
		JoinerAddress: raft.ServerAddress(in.JoinerAddress),
	})
	if errors.Is(err, raft.ErrNodeExist) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.Is(err, raft.ErrIsNotLeader) {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.JoinOut{}, nil
}
