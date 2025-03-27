package raft

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/raft"
	"time"
)

//todo use VerifyLeader for prevent stale read in store.Get

type Store struct {
	kvstore
	raft *raft.Raft
}

func NewStore(store kvstore, raft *raft.Raft) *Store {
	return &Store{
		kvstore: store,
		raft:    raft,
	}
}

func (s *Store) Put(ctx context.Context, key string, value string) error {
	if s.raft.VerifyLeader().Error() != nil {
		return ErrIsNotLeader
	}

	if err := s.kvstore.Put(ctx, key, value); err != nil {
		return err
	}

	cmd := command{
		Op:    opPut,
		Key:   key,
		Value: value,
	}

	return s.apply(ctx, cmd)
}

func (s *Store) Delete(ctx context.Context, key string) error {
	if s.raft.VerifyLeader().Error() != nil {
		return ErrIsNotLeader
	}

	if err := s.kvstore.Delete(ctx, key); err != nil {
		return err
	}

	cmd := command{
		Op:  opDelete,
		Key: key,
	}

	return s.apply(ctx, cmd)
}

func (s *Store) apply(ctx context.Context, cmd command) error {
	bytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	var timeout time.Duration

	deadline, ok := ctx.Deadline()
	if ok {
		timeout = time.Until(deadline)
	}

	return s.raft.Apply(bytes, timeout).Error()
}
