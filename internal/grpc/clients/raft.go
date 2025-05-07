package clients

import (
	"context"
	"errors"
	pb "github.com/HSE-RDBMS-course-work/kvstore-proto/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kvstore/internal/raft"
)

var ErrAddressIsEmpty = errors.New("address is nil")

type RaftClient struct {
	client pb.RaftClient
}

func NewRaftClient(addr string) (*RaftClient, error) {
	if addr == "" {
		return nil, ErrAddressIsEmpty
	}

	opts := []grpc.DialOption{ //todo
		grpc.WithTransportCredentials(insecure.NewCredentials()), //todo
	}

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}

	return &RaftClient{
		client: pb.NewRaftClient(conn),
	}, nil
}

func (rc *RaftClient) JoinToCluster(ctx context.Context, in raft.JoinToClusterIn) error {
	_, err := rc.client.JoinToCluster(ctx, &pb.JoinIn{
		JoinerId:      string(in.JoinerAddress),
		JoinerAddress: string(in.JoinerAddress),
	})

	return err
}
