package grpc

import (
	"context"
	"errors"
	pb "github.com/HSE-RDBMS-course-work/kvstore-proto/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"kvstore/internal/core"
	"kvstore/internal/raft"
)

type kvstore interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

type KVStoreServer struct {
	pb.UnimplementedKVStoreServer
	store kvstore
}

func NewKVStoreServer(store kvstore) *KVStoreServer {
	return &KVStoreServer{
		store: store,
	}
}

func (s *KVStoreServer) RegisterTo(server *grpc.Server) {
	pb.RegisterKVStoreServer(server, s)
}

func (s *KVStoreServer) Get(ctx context.Context, in *pb.GetIn) (*pb.GetOut, error) {
	key := in.GetKey()

	value, err := s.store.Get(ctx, key)
	if errors.Is(err, core.ErrNoKey) {
		return nil, status.Errorf(codes.NotFound, "there is no %s", key)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get value")
	}

	out := pb.GetOut{
		Value: value,
	}

	return &out, nil
}

func (s *KVStoreServer) Put(ctx context.Context, in *pb.PutIn) (*emptypb.Empty, error) {
	key := in.GetKey()
	value := in.GetValue()

	err := s.store.Put(ctx, key, value)
	if errors.Is(err, raft.ErrIsNotLeader) {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to put")
	}

	return empty, nil
}

func (s *KVStoreServer) Delete(ctx context.Context, in *pb.DeleteIn) (*emptypb.Empty, error) {
	key := in.GetKey()

	err := s.store.Delete(ctx, key)
	if errors.Is(err, raft.ErrIsNotLeader) {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete")
	}

	return empty, nil
}
