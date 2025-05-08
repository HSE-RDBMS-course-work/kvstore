package core

import (
	"context"
	"errors"
	"kvstore/internal/sl"
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
	CleanInterval   time.Duration
	CleanDuration   time.Duration
	InitialCapacity int64
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

	logger = logger.With(sl.Component("core.Store"))

	logger.Debug("creating new store", sl.Conf(conf))

	if conf.CleanInterval <= 0 {
		conf.CleanInterval = infinity
	}
	if conf.CleanDuration <= 0 {
		conf.CleanDuration = infinity
	}
	if conf.InitialCapacity <= 0 {
		conf.InitialCapacity = defaultCap
	}

	if conf.CleanInterval < conf.CleanDuration {
		return nil, errors.New("clean_interval cannot be less than clean_duration")
	}

	logger.Debug("created successfully", sl.Conf(conf))

	return &Store{
		expirations:      make(map[Key]time.Time, conf.InitialCapacity),
		mp:               make(map[Key]Value, conf.InitialCapacity),
		mu:               new(sync.RWMutex),
		logger:           logger,
		cleanInterval:    conf.CleanInterval,
		maxCleanDuration: conf.CleanDuration,
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
	} else {
		delete(s.expirations, key)
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

	s.logger.Info("start producing expired keys job")

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("producing expired keys was cancelled by context")
				return
			case <-timer.C:
				s.clean(ctx, ch)
			}
		}
	}()

	return ch
}

func (s *Store) clean(ctx context.Context, res chan<- Key) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, s.maxCleanDuration)
	defer cancel()

	s.logger.Debug("start cleaning interval", slog.Duration("duration", s.maxCleanDuration))
	defer func() {
		s.logger.Debug("end cleaning interval", slog.Duration("duration", s.maxCleanDuration))
	}()

	for k, expiration := range s.expirations {
		if !time.Now().After(expiration) {
			continue
		}

		select {
		case <-ctx.Done():
			return
		case res <- k:
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
