package core

import (
	"context"
	"errors"
	"log/slog"
	"maps"
	"sync"
	"time"
)

var (
	ErrNoKey = errors.New("error no key")
)

type Key string

type Value string

const (
	infinity   = time.Hour * 24 * 30 * 365
	defaultCap = 10_000
)

type Config struct {
	CleanInterval    time.Duration
	MaxCleanDuration time.Duration
	InitialCapacity  int64
}

type Store struct {
	expirations      map[Key]time.Time
	mp               map[Key]Value
	mu               *sync.RWMutex
	logger           *slog.Logger
	cleanInterval    time.Duration
	maxCleanDuration time.Duration
}

func NewStore(logger *slog.Logger, conf Config) (*Store, error) {
	if logger == nil {
		return nil, errors.New("logger required")
	}
	if conf.CleanInterval <= 0 {
		conf.CleanInterval = infinity
	}
	if conf.MaxCleanDuration <= 0 {
		conf.MaxCleanDuration = infinity
	}
	if conf.InitialCapacity <= 0 {
		conf.InitialCapacity = defaultCap
	}

	return &Store{
		expirations:      make(map[Key]time.Time, conf.InitialCapacity),
		mp:               make(map[Key]Value, conf.InitialCapacity),
		mu:               new(sync.RWMutex),
		logger:           logger,
		cleanInterval:    conf.CleanInterval,
		maxCleanDuration: conf.MaxCleanDuration,
	}, nil
}

func (s *Store) Get(_ context.Context, key Key) (*Value, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.mp[key]
	if !ok {
		return nil, ErrNoKey
	}

	expiration, ok := s.expirations[key]
	if ok && expiration.Before(time.Now()) {
		return nil, ErrNoKey
	}

	return &value, nil
}

func (s *Store) Put(_ context.Context, key Key, value Value, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mp[key] = value

	if ttl > 0 {
		s.expirations[key] = time.Now().Add(ttl)
	}

	return nil
}

func (s *Store) Delete(_ context.Context, key Key) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.mp, key)
	delete(s.expirations, key)

	return nil
}

func (s *Store) Expired(ctx context.Context) <-chan Key {
	ch := make(chan Key)
	timer := time.NewTicker(s.cleanInterval)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				s.clean(ctx, ch)
			}
		}
	}()

	return ch
}

func (s *Store) clean(ctx context.Context, res chan<- Key) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, s.maxCleanDuration)
	defer cancel()

	for k, expiration := range s.expirations {
		if !time.Now().After(expiration) {
			continue
		}

		select {
		case res <- k:
		case <-ctx.Done():
			return
		}
	}
}

func (s *Store) Snapshot(_ context.Context) (Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snap := Snapshot{
		Expirations: maps.Clone(s.expirations),
		Mp:          maps.Clone(s.mp),
	}

	return snap, nil
}

func (s *Store) Load(_ context.Context, snap Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mp = snap.Mp
	s.expirations = snap.Expirations

	return nil
}
