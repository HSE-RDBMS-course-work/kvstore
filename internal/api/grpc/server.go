package grpc

import (
	"context"
	"errors"
	pb "github.com/HSE-RDBMS-course-work/kvstore-proto/gen/go"
	"github.com/HSE-RDBMS-course-work/kvstore/internal/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	empty = &emptypb.Empty{}
)

type KVStore interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

type Server struct {
	pb.UnimplementedKVStoreServer
	store KVStore
}

func NewServer(store KVStore) *Server {
	return &Server{
		store: store,
	}
}

func (s *Server) RegisterTo(server *grpc.Server) {
	pb.RegisterKVStoreServer(server, s)
}

func (s *Server) Get(ctx context.Context, in *pb.GetIn) (*pb.GetOut, error) {
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

func (s *Server) Put(ctx context.Context, in *pb.PutIn) (*emptypb.Empty, error) {
	key := in.GetKey()
	value := in.GetValue()

	err := s.store.Put(ctx, key, value)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to put")
	}

	return empty, nil
}

func (s *Server) Delete(ctx context.Context, in *pb.DeleteIn) (*emptypb.Empty, error) {
	key := in.GetKey()

	err := s.store.Delete(ctx, key)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete")
	}

	return empty, nil
}
