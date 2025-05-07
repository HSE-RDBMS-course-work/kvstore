package raft

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/raft"
	"io"
	"kvstore/internal/sl"
	"log/slog"
)

// FSM is an implementation of final state machine
// it is used by raft to apply logs from leader or from snapshots to store
type FSM struct {
	logger *slog.Logger
	store  kvstore
}

func NewFSM(logger *slog.Logger, store kvstore) (*FSM, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}

	logger = logger.With(sl.Component("raft.FSM"))

	if store == nil {
		return nil, errors.New("store required")
	}

	logger.Debug("created successfully")

	return &FSM{
		logger: logger,
		store:  store,
	}, nil
}

// Apply applies logs from leader to replicas
// for more info check docs for raft.FSM
func (fsm *FSM) Apply(log *raft.Log) any {
	var cmd command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		fsm.logger.Warn("got incorrect json with command", sl.Error(err))
		return err
	}

	fsm.logger.Info("applying command", cmd)

	var err error
	switch cmd.Op {
	case opPut:
		err = fsm.store.Put(context.Background(), cmd.Key, cmd.Value, cmd.TTL)
	case opDelete:
		err = fsm.store.Delete(context.Background(), cmd.Key)
	default:
		err = ErrUnknownCmd
	}

	if err != nil {
		return err
	}

	return nil
}

func (fsm *FSM) Snapshot() (raft.FSMSnapshot, error) {
	snap, err := fsm.store.Snapshot(context.Background())
	if err != nil {
		return nil, err
	}

	return &snapshot{
		Snapshot: snap,
	}, nil
}

func (fsm *FSM) Restore(reader io.ReadCloser) error {
	var snap snapshot
	if err := json.NewDecoder(reader).Decode(&snap); err != nil {
		return err
	}

	if err := fsm.store.Load(context.Background(), snap.Snapshot); err != nil {
		return err
	}

	return nil
}
