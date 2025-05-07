package raft

import (
	"context"
	"errors"
	"github.com/hashicorp/raft"
	"kvstore/internal/sl"
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
	BootstrapCluster  bool
}

type ClusterNode struct {
	logger            *slog.Logger
	raft              *raft.Raft
	existLeader       existLeader
	id                ServerID
	realAddress       ServerAddress
	advertisedAddress ServerAddress
	isFirstNode       bool
}

func NewClusterNode(logger *slog.Logger, r *raft.Raft, existLeader existLeader, conf ClusterNodeConfig) (*ClusterNode, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	logger = logger.With(sl.Component("raft.ClusterNode"))

	if r == nil {
		return nil, errors.New("raft instance required")
	}

	logger.Debug("creating cluster node", sl.Conf(conf))

	if existLeader == nil && conf.BootstrapCluster {
		return nil, errors.New("existLeaderClient is required if you bootstrap the cluster")
	}
	if conf.RealAddress == "" {
		return nil, errors.New("real address required")
	}
	if conf.AdvertisedAddress == "" {
		return nil, errors.New("advertised address required")
	}
	if conf.ID == "" {
		return nil, errors.New("nodeID required")
	}

	logger.Debug("created successfully", sl.Conf(conf))

	return &ClusterNode{
		logger:            logger,
		raft:              r,
		existLeader:       existLeader,
		id:                conf.ID,
		realAddress:       conf.RealAddress,
		advertisedAddress: conf.AdvertisedAddress,
		isFirstNode:       conf.BootstrapCluster,
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
	r.logger.Info("starting listening", slog.String("address", string(r.realAddress)))

	if r.isFirstNode {
		return r.bootstrapCluster(ctx)
	}

	return r.joinToCluster(ctx)
}

func (r *ClusterNode) Shutdown() error {
	r.logger.Info("shutting down")
	if err := r.raft.Shutdown().Error(); err != nil {
		return err
	}
	r.logger.Info("shut down gracefully")
	return nil
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
