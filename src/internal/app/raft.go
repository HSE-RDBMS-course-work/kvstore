package app

import (
	"context"
	"fmt"
	pb "github.com/HSE-RDBMS-course-work/kvstore-proto/gen/go"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"kvstore/internal/config"
	"log"
	"net"
	"os"
)

func StartRaftNode(cfg *config.RaftConfig, fsm raft.FSM) *raft.Raft {
	raftConfig := raft.DefaultConfig() //todo pass logger
	raftConfig.LocalID = raft.ServerID(cfg.LocalID)

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

	localAddr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	advertisedAddr, err := net.ResolveTCPAddr("tcp", cfg.Advertise)
	if err != nil {
		log.Fatalf("cannot resolve advertised address: %v", err)
	}

	transport, err := raft.NewTCPTransport(localAddr, advertisedAddr, cfg.MaxPool, cfg.Timeout, os.Stderr)
	if err != nil {
		log.Fatalf("cannot create transport: %v", err)
	}

	node, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		log.Fatalf("cannot create raft node: %v", err)
	}

	if cfg.LeaderAddr == "" {
		bootstrap(node, cfg.LocalID, localAddr)
	} else {
		if err := join(cfg.LeaderAddr, cfg.LocalID, cfg.Advertise); err != nil {
			log.Print(err) //todo
		}
	}

	return node
}

func bootstrap(leader *raft.Raft, id string, addr string) {
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(id),
				Address: raft.ServerAddress(addr),
			},
		},
	}

	go func() {
		//todo catch error
		leader.BootstrapCluster(configuration)
	}()
}

func join(leaderAddr, id, addr string) error {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), //todo
	}

	conn, err := grpc.NewClient(leaderAddr, opts...)
	if err != nil {
		return err
	}

	client := pb.NewRaftClient(conn)

	in := pb.JoinIn{
		NodeAddr: addr,
		NodeID:   id,
	}

	_, err = client.Join(context.TODO(), &in)
	if err != nil {
		return err
	}

	return nil
}
