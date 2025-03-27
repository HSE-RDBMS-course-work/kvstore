package grpc

import (
	"context"
	"errors"
	pb "github.com/HSE-RDBMS-course-work/kvstore-proto/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"kvstore/internal/raft"
)

type cluster interface {
	AcceptJoin(ctx context.Context, addr string, id string) error
}

type RaftServer struct {
	pb.UnimplementedRaftServer
	cluster cluster
}

func NewRaftServer(cluster cluster) *RaftServer {
	return &RaftServer{
		cluster: cluster,
	}
}

func (s *RaftServer) RegisterTo(server *grpc.Server) {
	pb.RegisterRaftServer(server, s)
}

func (s *RaftServer) Join(ctx context.Context, in *pb.JoinIn) (*emptypb.Empty, error) {
	err := s.cluster.AcceptJoin(ctx, in.GetNodeAddr(), in.GetNodeID())
	if errors.Is(err, raft.ErrNodeExist) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if errors.Is(err, raft.ErrIsNotLeader) {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if err != nil { //returns it only for auth users so we can return detailed msg
		return nil, status.Error(codes.Internal, err.Error())
	}

	return empty, nil
}
