package template

import (
	"context"

	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

type bufferedSequencerUpdate struct {
	operation Operation
	visPrefix int64
	cid       int64
}

type BufferedSequencer struct {
	role      util.Role
	dataType  ReplicatedDataType
	confirmed []Operation
	peers     map[int64]network.Link
}

func (s *BufferedSequencer) Perform(op Operation) OperationResult {
	if s.dataType.IsReadOnly(op) {
		return s.dataType.ComputeResult(op, s.confirmed)
	} else if s.dataType.IsUpdateOnly(op) {
		for _, c := range s.peers {
			c.Send(context.Background(), op)
		}
		return true
	}

	return nil
}
