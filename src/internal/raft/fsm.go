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

var (
	ErrUnknownCmd = errors.New("error unknown command")
)

type FSM struct {
	log   *slog.Logger
	store kvstore
}

func NewFSM(store kvstore, log *slog.Logger) *FSM {
	return &FSM{
		store: store,
		log:   log,
	}
}

// Apply applies logs from leader to replicas
// for more info check docs for raft.FSM
func (fsm *FSM) Apply(log *raft.Log) any {
	var cmd command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		fsm.log.Warn("got incorrect json with command", sl.Error(err))
		return err
	}

	fsm.log.Info("applying command", slCommand(cmd))

	var err error
	switch cmd.Op {
	case opPut:
		err = fsm.store.Put(context.Background(), cmd.Key, cmd.Value)
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
	mp, err := fsm.store.Data(context.Background())
	if err != nil {
		return nil, err
	}

	return &snapshot{mp: mp}, nil
}

func (fsm *FSM) Restore(snapshot io.ReadCloser) error {
	var mp map[string]string
	if err := json.NewDecoder(snapshot).Decode(&mp); err != nil { //todo add sync.Pool
		return err
	}

	if err := fsm.store.Load(context.Background(), mp); err != nil {
		return err
	}

	return nil
}
