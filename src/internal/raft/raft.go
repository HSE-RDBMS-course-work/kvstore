package raft

import (
	"fmt"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"io"
	"net"
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

func New(fsm raft.FSM, out io.Writer, conf Config) (*raft.Raft, error) {
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = ServerID(conf.NodeID)

	logStore, err := raftboltdb.NewBoltStore(conf.DataLocation)
	if err != nil {
		return nil, fmt.Errorf("cannnot create raft logger store: %v", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(conf.DataLocation)
	if err != nil {
		return nil, fmt.Errorf("cannot create raft stable store: %v", err)
	}

	snapshots, err := raft.NewFileSnapshotStore(conf.DataLocation, conf.SnapshotsRetain, out)
	if err != nil {
		return nil, fmt.Errorf("cannot create snapshot store: %v", err)
	}

	advertisedAddr, err := net.ResolveTCPAddr("tcp", conf.AdvertisedAddress)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve advertised address: %v", err)
	}

	transport, err := raft.NewTCPTransport(conf.RealAddress, advertisedAddr, conf.MaxPool, conf.Timeout, out)
	if err != nil {
		return nil, fmt.Errorf("cannot create transport: %v", err)
	}

	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("cannot create raft.Raft r: %v", err)
	}

	return r, nil
}
