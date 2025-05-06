package raft

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/hashicorp/raft"
	"log/slog"
)

type JoinToClusterIn struct {
	JoinerID      ServerID
	JoinerAddress ServerAddress
}

type existLeader interface {
	JoinToCluster(context context.Context, in JoinToClusterIn) error
}

type ClusterNodeConfig struct {
	ID                ServerID
	RealAddress       ServerAddress
	AdvertisedAddress ServerAddress
}

type ClusterNode struct {
	raft              *raft.Raft
	existLeader       existLeader
	id                ServerID
	realAddress       ServerAddress
	advertisedAddress ServerAddress
}

func NewClusterNode(logger *slog.Logger, r *raft.Raft, existLeader existLeader, conf ClusterNodeConfig) (*ClusterNode, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}
	if r == nil {
		return nil, errors.New("raft instance required")
	}
	if conf.RealAddress == "" {
		return nil, errors.New("RealAddress required")
	}
	if conf.AdvertisedAddress == "" {
		return nil, errors.New("AdvertisedAddress required")
	}

	if conf.ID == "" {
		conf.ID = ServerID(uuid.New().String()) //todo так реально можно?
	}

	return &ClusterNode{
		raft:              r,
		existLeader:       existLeader,
		id:                conf.ID,
		realAddress:       conf.RealAddress,
		advertisedAddress: conf.AdvertisedAddress,
	}, nil
}

func (r *ClusterNode) AcceptJoin(ctx context.Context, in JoinToClusterIn) error {
	if r.raft.State() != raft.Leader { //todo redirect it to real existLeader
		return ErrIsNotLeader
	}

	conf := r.raft.GetConfiguration()
	if err := conf.Error(); err != nil {
		return err
	}

	for _, srv := range conf.Configuration().Servers {
		if srv.Address == in.JoinerAddress || srv.ID == in.JoinerID {
			return newErrNodeExist(in.JoinerAddress, in.JoinerID)
		}
	}

	addVoter := r.raft.AddVoter(in.JoinerID, in.JoinerAddress, 0, 0)
	if addVoter.Error() != nil {
		return addVoter.Error()
	}

	return nil
}

func (r *ClusterNode) Run(ctx context.Context) error {
	if r.existLeader != nil {
		return r.bootstrapCluster(ctx)
	}

	return r.joinToCluster(ctx)
}

func (r *ClusterNode) Shutdown() error {
	return r.raft.Shutdown().Error()
}

func (r *ClusterNode) bootstrapCluster(ctx context.Context) error {
	future := r.raft.BootstrapCluster(raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      r.id,
				Address: r.realAddress,
			},
		},
	})

	return future.Error() //todo use context to catch timeout
}

func (r *ClusterNode) joinToCluster(ctx context.Context) error {
	return r.existLeader.JoinToCluster(ctx, JoinToClusterIn{
		JoinerID:      r.id,
		JoinerAddress: r.advertisedAddress,
	})
}
