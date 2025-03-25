package fsm

import (
	"encoding/json"
	"errors"
	"github.com/hashicorp/raft"
)

type Snapshot struct {
	mp map[string]string
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) (err error) {
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

	data, err := json.Marshal(s.mp)
	if err != nil {
		return err
	}
	if _, err := sink.Write(data); err != nil {
		return err
	}
	return nil
}

func (s *Snapshot) Release() {}
