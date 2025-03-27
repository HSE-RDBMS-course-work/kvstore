package raft

import (
	"context"
	"github.com/hashicorp/raft"
)

type Server struct {
	raft *raft.Raft
}

func NewServer(raft *raft.Raft) *Server {
	return &Server{
		raft: raft,
	}
}

//todo refactor

func (s *Store) AcceptJoin(ctx context.Context, nodeAddr string, nodeID string) error {
	if s.raft.VerifyLeader().Error() != nil {
		return ErrIsNotLeader
	}

	cfg := s.raft.GetConfiguration()
	if err := cfg.Error(); err != nil {
		return err
	}

	id := raft.ServerID(nodeID)
	addr := raft.ServerAddress(nodeAddr)

	for _, srv := range cfg.Configuration().Servers {
		if srv.Address == addr && srv.ID == id {
			return newErrNodeExist(nodeAddr, nodeID)
		}
	}

	addVoter := s.raft.AddVoter(id, addr, 0, 0)
	if addVoter.Error() != nil {
		return addVoter.Error()
	}

	return nil
}
