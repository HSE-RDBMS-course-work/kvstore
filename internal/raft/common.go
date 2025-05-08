package raft

import (
	"context"
	"github.com/hashicorp/raft"
	"kvstore/internal/core"
	"log/slog"
	"time"
)

type ServerID = raft.ServerID

type ServerAddress = raft.ServerAddress

type kvstore interface {
	Get(context.Context, core.Key) (*core.Value, error)
	Put(context.Context, core.Key, core.Value, time.Duration) error
	Delete(context.Context, core.Key) error
	Expired(context.Context) <-chan core.Key
	Snapshot(context.Context) (core.Snapshot, error)
	Load(context.Context, core.Snapshot) error
}

type operation string

const (
	opPut    operation = "put"
	opDelete operation = "delete"
)

type command struct {
	Op    operation     `json:"op"`
	Key   core.Key      `json:"key"`
	Value core.Value    `json:"value"`
	TTL   time.Duration `json:"ttl"`
}

func (cmd *command) LogAttr() slog.Attr {
	return slog.Group(
		"command",
		slog.String("op", string(cmd.Op)),
		slog.String("key", string(cmd.Key)),
		slog.String("value", string(cmd.Value)),
		slog.Duration("ttl", cmd.TTL),
	)
}
