package raft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"kvstore/internal/core"
	"kvstore/internal/sl"
	"log/slog"
	"time"
)

// Store make some key value storage distributed via raft
type Store struct {
	logger *slog.Logger
	raft   *raft.Raft
	store  kvstore
}

func NewStore(logger *slog.Logger, raft *raft.Raft, store kvstore) (*Store, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	logger = logger.With(sl.Component("raft.Store"))

	if raft == nil {
		return nil, errors.New("raft required")
	}
	if store == nil {
		return nil, errors.New("store required")
	}

	logger.Debug("created successfully")

	return &Store{
		logger: logger,
		raft:   raft,
		store:  store,
	}, nil
}

func (s *Store) Get(ctx context.Context, key core.Key) (*core.Value, error) {
	return s.store.Get(ctx, key)
}

func (s *Store) ConsistentGet(ctx context.Context, key core.Key) (*core.Value, error) {
	if s.raft.State() != raft.Leader {
		return nil, newErrorIsNotLeader(s.raft)
	}

	return s.store.Get(ctx, key)
}

func (s *Store) Put(ctx context.Context, key core.Key, value core.Value, ttl time.Duration) error {
	if s.raft.State() != raft.Leader {
		return newErrorIsNotLeader(s.raft)
	}

	return s.apply(ctx, command{
		Op:    opPut,
		Key:   key,
		Value: value,
		TTL:   ttl,
	})
}

func (s *Store) Delete(ctx context.Context, key core.Key) error {
	if s.raft.VerifyLeader().Error() != nil {
		return newErrorIsNotLeader(s.raft)
	}

	return s.apply(ctx, command{
		Op:  opDelete,
		Key: key,
	})
}

func (s *Store) RunCleaning(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case isLeader, ok := <-s.raft.LeaderCh():
			if !ok {
				return errors.New("raft leader channel closed")
			}
			if !isLeader {
				continue
			}
		}

		for key := range s.store.Expired(ctx) {
			if err := s.Delete(ctx, key); err != nil {
				s.logger.Warn("failed to delete expired key")
			}
		}
	}
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

	err = s.raft.Apply(bytes, timeout).Error()
	if err != nil {
		return fmt.Errorf("appling log to other nodes: %w", err)
	}

	s.logger.Debug("applied command", cmd.LogAttr())

	return nil
}
