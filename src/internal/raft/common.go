package raft

import (
	"context"
	"log/slog"
)

type kvstore interface {
	Data(ctx context.Context) (map[string]string, error)
	Get(ctx context.Context, key string) (string, error)
	Load(ctx context.Context, mp map[string]string) error
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

type operation string

const (
	opPut    = "put"
	opDelete = "delete"
)

type command struct {
	Op    operation `json:"op"`
	Key   string    `json:"key"`
	Value string    `json:"value"`
}

func slCommand(cmd command) slog.Attr {
	return slog.Group(
		"command",
		slog.String("op", string(cmd.Op)),
		slog.String("key", cmd.Key),
		slog.String("value", cmd.Value),
	)
}
