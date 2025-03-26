package raft

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/raft"
	"io"
)

var (
	ErrUnknownCmd = errors.New("error unknown command")
)

type FSM struct {
	store kvstore
}

func NewFSM(store kvstore) *FSM {
	return &FSM{
		store: store,
	}
}

func (fsm *FSM) Apply(log *raft.Log) any {
	var cmd command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}

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
	if err := json.NewDecoder(snapshot).Decode(&mp); err != nil {
		return err
	}

	for k, v := range mp {
		//todo no lock required according to hashicorp docs
		fsm.store.Put(context.Background(), k, v) //todo fill store from reader
	}

	return nil
}
