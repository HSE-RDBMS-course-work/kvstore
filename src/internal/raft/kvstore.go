package raft

import (
	"context"
)

type kvstore interface {
	Data(ctx context.Context) (map[string]string, error)
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}
