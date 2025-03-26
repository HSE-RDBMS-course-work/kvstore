package raft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"log"
	"time"
)

var ErrNodeExist = errors.New("error node exists")

type errNodeExist struct {
	id   string
	addr string
}

func newErrNodeExist(addr string, id string) error {
	return &errNodeExist{
		addr: addr,
		id:   id,
	}
}

func (e *errNodeExist) Error() string {
	return fmt.Sprintf(
		"node with nodeID - %s or at - %s already member of cluster, ignoring join request",
		e.id,
		e.addr,
	)
}

func (e *errNodeExist) Unwrap() error {
	return ErrNodeExist
}

type Node struct {
	kvstore
	raft *raft.Raft
}

func NewNode(store kvstore, raft *raft.Raft) *Node {
	return &Node{
		kvstore: store,
		raft:    raft,
	}
}

func (s *Node) Put(ctx context.Context, key, value string) error {
	if err := s.raft.VerifyLeader().Error(); err != nil { //todo catch this error type above
		return err
	}

	if err := s.kvstore.Put(ctx, key, value); err != nil {
		return err
	}

	cmd := command{
		Op:    opPut,
		Key:   key,
		Value: value,
	}

	s.apply(ctx, cmd)

	return nil
}

func (s *Node) Delete(ctx context.Context, key string) error {
	if err := s.raft.VerifyLeader().Error(); err != nil {
		return err
	}

	if err := s.kvstore.Delete(ctx, key); err != nil {
		return err
	}

	cmd := command{
		Op:  opDelete,
		Key: key,
	}

	s.apply(ctx, cmd)

	return nil
}

func (s *Node) AcceptJoin(ctx context.Context, nodeAddr string, nodeID string) error {
	//todo maybe move this method from here

	log.Printf("got join request: %s %s\n", nodeAddr, nodeID)

	cfg := s.raft.GetConfiguration()
	if err := cfg.Error(); err != nil {
		return err
	}

	id := raft.ServerID(nodeID)
	addr := raft.ServerAddress(nodeAddr)

	for _, srv := range cfg.Configuration().Servers {
		if srv.Address == addr || srv.ID == id {
			return newErrNodeExist(nodeAddr, nodeID)
		}
	}

	addVoter := s.raft.AddVoter(id, addr, 0, 0)
	if addVoter.Error() != nil {
		return addVoter.Error()
	}

	return nil
}

func (s *Node) apply(ctx context.Context, cmd command) {
	bytes, err := json.Marshal(cmd)
	if err != nil {
		//todo log it
	}

	var timeout time.Duration

	deadline, ok := ctx.Deadline()
	if ok {
		timeout = time.Until(deadline)
	}

	//todo apply can be used only on leader
	//todo test it
	//todo think about consistent read only from leader and inconsistent read from follower
	s.raft.Apply(bytes, timeout)
}
