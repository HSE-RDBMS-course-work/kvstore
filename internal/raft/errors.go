package raft

import (
	"errors"
)

var (
	ErrIsNotLeader = errors.New("this node is not a leader")
	ErrUnknownCmd  = errors.New("unknown command")
)

type ErrorIsNotLeader struct {
	err           error
	leaderAddress string
}

func newErrorIsNotLeader(leaderAddress string) *ErrorIsNotLeader {
	return &ErrorIsNotLeader{
		err:           ErrIsNotLeader,
		leaderAddress: leaderAddress,
	}
}

func (e *ErrorIsNotLeader) Error() string {
	return e.err.Error()
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
