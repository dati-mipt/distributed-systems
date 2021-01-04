package template

import (
	"context"
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

type AsyncSequencer struct {
	role      util.Role
	dataType  ReplicatedDataType
	confirmed []Operation
	peers     map[int64]network.Link
}

func NewAsyncSequencer(role util.Role, dataType ReplicatedDataType) *AsyncSequencer {
	return &AsyncSequencer{
		role:     role,
		dataType: dataType,
		peers:    map[int64]network.Link{},
	}
}

func (s *AsyncSequencer) Perform(op Operation) OperationResult {
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

func (s *AsyncSequencer) Introduce(cid int64, link network.Link) {
	if link == nil {
		return
	}

	switch s.role {
	case util.Client:
		if _, ok := s.peers[0]; !ok && cid == 0 {
			s.peers[cid] = link
		}
	case util.Server:
		if cid != 0 {
			s.peers[cid] = link
		}
	}
}

func (s *AsyncSequencer) Receive(rid int64, msg interface{}) interface{} {
	if op, ok := msg.(Operation); ok {
		s.confirmed = append(s.confirmed, op)
		if s.role == util.Server {
			for _, c := range s.peers {
				c.Send(context.Background(), op)
			}
		}
	}

	return nil
}
