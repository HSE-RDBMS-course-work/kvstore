package raft

import (
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
)

var (
	ErrIsNotLeader = errors.New("this node is not a leader")
	ErrUnknownCmd  = errors.New("unknown command")
)

type ErrorIsNotLeader struct {
	err           error
	leaderAddress string
}

func newErrorIsNotLeader(r *raft.Raft) *ErrorIsNotLeader {
	_, leaderPublicAddress := r.LeaderWithID()

	return &ErrorIsNotLeader{
		err:           ErrIsNotLeader,
		leaderAddress: string(leaderPublicAddress),
	}
}

func (e *ErrorIsNotLeader) Error() string {
	return fmt.Sprintf("leader=%s, this node is not a leader", e.Leader())
}

func (e *ErrorIsNotLeader) Unwrap() error {
	return e.err
}

func (e *ErrorIsNotLeader) Leader() string {
	if e.leaderAddress == "" {
		return "leader was not specified"
	}
	return e.leaderAddress
}
