package servers

import (
	"context"
	"errors"
	pb "github.com/HSE-RDBMS-course-work/kvstore-proto/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kvstore/internal/core"
	"kvstore/internal/raft"
	"time"
)

type getFn = func(context.Context, core.Key) (*core.Value, error)

type kvstore interface {
	Get(ctx context.Context, key core.Key) (*core.Value, error)
	ConsistentGet(ctx context.Context, key core.Key) (*core.Value, error)
	Put(ctx context.Context, key core.Key, value core.Value, ttl time.Duration) error
	Delete(ctx context.Context, key core.Key) error
}

type KVStoreServer struct {
	pb.UnimplementedKVStoreServer
	store kvstore
}

func NewKVStoreServer(store kvstore) (*KVStoreServer, error) {
	if store == nil {
		return nil, errors.New("store is required")
	}

	return &KVStoreServer{
		store: store,
	}, nil
}

func (s *KVStoreServer) RegisterTo(server *grpc.Server) {
	pb.RegisterKVStoreServer(server, s)
}

func (s *KVStoreServer) Get(ctx context.Context, in *pb.GetIn) (*pb.GetOut, error) {
	return s.get(ctx, in, s.store.Get)
}

func (s *KVStoreServer) ConsistentGet(ctx context.Context, in *pb.GetIn) (*pb.GetOut, error) {
	return s.get(ctx, in, s.store.ConsistentGet)
}

func (s *KVStoreServer) Put(ctx context.Context, in *pb.PutIn) (*pb.PutOut, error) {
	key := core.Key(in.GetKey())
	value := core.Value(in.GetValue())
	ttl := time.Duration(in.GetTtl())

	err := s.store.Put(ctx, key, value, ttl)
	if errors.Is(err, raft.ErrIsNotLeader) {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to put")
	}

	return &pb.PutOut{}, nil
}

func (s *KVStoreServer) Delete(ctx context.Context, in *pb.DeleteIn) (*pb.DeleteOut, error) {
	key := core.Key(in.GetKey())

	err := s.store.Delete(ctx, key)
	if errors.Is(err, raft.ErrIsNotLeader) {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete")
	}

	return &pb.DeleteOut{}, nil
}

func (s *KVStoreServer) get(ctx context.Context, in *pb.GetIn, get getFn) (*pb.GetOut, error) {
	key := core.Key(in.GetKey())

	value, err := get(ctx, key)
	if errors.Is(err, core.ErrNoKey) {
		return nil, status.Errorf(codes.NotFound, "there is no %s", key)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get value")
	}

	var outValue string
	if value != nil {
		outValue = string(*value)
	}

	out := pb.GetOut{
		Value: outValue,
	}

	return &out, nil
}
