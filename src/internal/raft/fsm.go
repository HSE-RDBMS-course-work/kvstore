package fsm

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

type Command struct {
	Op    string `json:"op"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type kvstore interface {
	Map(ctx context.Context) (map[string]string, error)
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

type FSM struct {
	store kvstore
}

func NewFSM(store kvstore) *FSM {
	return &FSM{
		store: store,
	}
}

func (fsm *FSM) Apply(log *raft.Log) any {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}

	var err error
	switch cmd.Op {
	case "put":
		err = fsm.store.Put(context.Background(), cmd.Key, cmd.Value)
	case "delete":
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
	mp, err := fsm.store.Map(context.Background())
	if err != nil {
		return nil, err
	}

	return &Snapshot{mp: mp}, nil
}

func (fsm *FSM) Restore(snapshot io.ReadCloser) error {
	var mp map[string]string
	if err := json.NewDecoder(snapshot).Decode(&mp); err != nil {
		return err
	}

	for k, v := range mp {
		fsm.store.Put(context.Background(), k, v) //todo
	}

	return nil
}
