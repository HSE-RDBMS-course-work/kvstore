package clients

import (
	"context"
	"errors"
	pb "github.com/HSE-RDBMS-course-work/kvstore-proto/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kvstore/internal/grpc/clients/interceptors"
	"kvstore/internal/raft"
)

var ErrAddressIsEmpty = errors.New("address is nil")

type RaftClientConfig struct {
	Address  string
	Username string
	Password string
}

type RaftClient struct {
	client pb.RaftClient
}

func NewRaftClient(conf RaftClientConfig) (*RaftClient, error) {
	if conf.Address == "" {
		return nil, ErrAddressIsEmpty
	}
	if conf.Username == "" {
		return nil, errors.New("username is required")
	}
	if conf.Password == "" {
		return nil, errors.New("password is required")
	}

	opts := []grpc.DialOption{ //todo
		grpc.WithTransportCredentials(insecure.NewCredentials()), //todo
		grpc.WithUnaryInterceptor(interceptors.NewAuth(conf.Username, conf.Password)),
	}

	conn, err := grpc.NewClient(conf.Address, opts...)
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
