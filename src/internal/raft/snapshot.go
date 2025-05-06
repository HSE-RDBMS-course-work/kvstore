package raft

import (
	"encoding/json"
	"errors"
	"github.com/hashicorp/raft"
	"kvstore/internal/core"
)

type snapshot struct {
	core.Snapshot
}

func (s *snapshot) Persist(sink raft.SnapshotSink) (err error) {
	defer func() {
		var closeFunc func() error
		if err != nil {
			closeFunc = sink.Cancel
		} else {
			closeFunc = sink.Close
		}

		if closeErr := closeFunc(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	data, err := json.Marshal(s.Snapshot)
	if err != nil {
		return err
	}
	if _, err := sink.Write(data); err != nil {
		return err
	}
	return nil
}

func (s *snapshot) Release() {}
