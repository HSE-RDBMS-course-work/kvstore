package core

import (
	"context"
	"errors"
	"maps"
	"sync"
)

var (
	ErrNoKey = errors.New("error no key")
)

type Store struct {
	mp map[string]string
	mu *sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		mp: make(map[string]string),
		mu: new(sync.RWMutex),
	}
}

func (s *Store) Data(ctx context.Context) (map[string]string, error) {
	return maps.Clone(s.mp), nil
}

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.mp[key]
	if !ok {
		return "", ErrNoKey
	}

	return value, nil
}

func (s *Store) Load(ctx context.Context, mp map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mp = mp

	return nil
}

func (s *Store) Put(ctx context.Context, key string, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mp[key] = value

	return nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.mp, key)

	return nil
}
