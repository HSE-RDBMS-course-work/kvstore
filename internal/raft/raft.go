package raft

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"kvstore/internal/sl"
	"log/slog"
	"net"
	"path/filepath"
	"time"
)

type Config struct {
	RealAddress       string
	AdvertisedAddress string
	NodeID            string
	DataLocation      string
	SnapshotsRetain   int
	MaxPool           int
	Timeout           time.Duration
}

func New(logger *slog.Logger, hcLogger hclog.Logger, fsm raft.FSM, conf Config) (*raft.Raft, bool, error) {
	logger = logger.With(sl.Component("raft.New"))
	logger.Debug("creating raft instance", sl.Conf(conf))

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = ServerID(conf.NodeID)
	raftConfig.Logger = hcLogger

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(conf.DataLocation, "log.db"))
	if err != nil {
		return nil, false, fmt.Errorf("cannnot create raft log store: %v", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(conf.DataLocation, "stable.db"))
	if err != nil {
		return nil, false, fmt.Errorf("cannot create raft stable store: %v", err)
	}

	snapshots, err := raft.NewFileSnapshotStoreWithLogger(conf.DataLocation, conf.SnapshotsRetain, hcLogger)
	if err != nil {
		return nil, false, fmt.Errorf("cannot create snapshot store: %v", err)
	}

	advertisedAddr, err := net.ResolveTCPAddr("tcp", conf.AdvertisedAddress)
	if err != nil {
		return nil, false, fmt.Errorf("cannot resolve advertised address: %v", err)
	}

	transport, err := raft.NewTCPTransportWithLogger(conf.RealAddress, advertisedAddr, conf.MaxPool, conf.Timeout, hcLogger)
	if err != nil {
		return nil, false, fmt.Errorf("cannot create transport: %v", err)
	}

	hasState, err := raft.HasExistingState(logStore, stableStore, snapshots)
	if err != nil {
		return nil, false, fmt.Errorf("cannot check existing state: %v", err)
	}

	if hasState { //todo это должно быть в другом месте
		err := raft.RecoverCluster(raftConfig, fsm, logStore, stableStore, snapshots, transport, raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      ServerID(conf.NodeID),
					Address: ServerAddress(conf.AdvertisedAddress),
				},
			},
		})
		if err != nil {
			return nil, false, fmt.Errorf("cannot recover cluster: %v", err)
		}
		logger.Debug("recovered cluster node successfully")
	}

	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, false, fmt.Errorf("cannot create raft.Raft r: %v", err)
	}

	logger.Debug("created successfully", sl.Conf(conf))

	return r, hasState, nil
}
