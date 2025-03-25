package app

import (
	"fmt"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"log"
	"os"
	"time"
)

type Node struct {
	Address string `yaml:"address"`
	ID      string `yaml:"id"`
}

type RaftConfig struct {
	Host             string `yaml:"host"`
	Port             string `yaml:"port"`
	LocalID          string `yaml:"local_id"`
	LogLocation      string `yaml:"log_location"`
	StableLocation   string `yaml:"stable_location"`
	SnapshotLocation string `yaml:"snapshot_location"`
	Nodes            []Node `yaml:"nodes"`
}

func StartRaftNode(cfg *RaftConfig, fsm raft.FSM) *raft.Raft {
	config := raft.DefaultConfig() //todo pass logger
	config.LocalID = raft.ServerID(cfg.LocalID)

	logStore, err := raftboltdb.NewBoltStore(cfg.LogLocation)
	if err != nil {
		log.Fatalf("cannnot create raft log store: %v", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(cfg.StableLocation)
	if err != nil {
		log.Fatalf("cannot create raft stable store: %v", err)
	}

	snapshots, err := raft.NewFileSnapshotStore(cfg.SnapshotLocation, 2, os.Stderr)
	if err != nil {
		log.Fatalf("cannot create snapshot store: %v", err)
	}

	nodeAddr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	//todo first is internal nodeAddr second is public nodeAddr
	transport, err := raft.NewTCPTransport(nodeAddr, nil, 3, 10*time.Second, os.Stderr)
	if err != nil {
		log.Fatalf("cannot create transport: %v", err)
	}

	node, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		log.Fatalf("cannot create raft node: %v", err)
	}

	nodes := make([]raft.Server, len(cfg.Nodes)+1)

	for i, node := range cfg.Nodes {
		nodes[i] = raft.Server{
			ID:      raft.ServerID(node.ID),
			Address: raft.ServerAddress(node.Address),
		}
	}

	nodes = append(nodes, raft.Server{ //appending local node
		ID:      raft.ServerID(cfg.LocalID),
		Address: raft.ServerAddress(nodeAddr),
	})

	configuration := raft.Configuration{
		Servers: nodes,
	}

	go func() {
		node.BootstrapCluster(configuration)
	}()

	return node
}
