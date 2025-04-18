package raft

import (
	"errors"
	"fmt"
)

var (
	ErrIsNotLeader = errors.New("this node is not a leader")
	ErrNodeExist   = errors.New("error node already in cluster")
	ErrUnknownCmd  = errors.New("error unknown command")
)

type errNodeExist struct {
	id      string
	address string
}

func newErrNodeExist(address string, id string) error {
	return &errNodeExist{
		address: address,
		id:      id,
	}
}

func (e *errNodeExist) Error() string {
	return fmt.Sprintf(
		"node with nodeID - %s or at - %s already member of cluster, ignoring join request",
		e.id,
		e.address,
	)
}

func (e *errNodeExist) Unwrap() error {
	return ErrNodeExist
}
