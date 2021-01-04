package template

import (
	"context"
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

type Sequencer struct {
	role      util.Role
	dataType  ReplicatedDataType
	confirmed []Operation
	peers     map[int64]network.Link
}

func NewSequentialServer(role util.Role, dataType ReplicatedDataType) *Sequencer {
	return &Sequencer{
		role:     role,
		dataType: dataType,
		peers:    map[int64]network.Link{},
	}
}

func (s *Sequencer) Perform(op Operation) OperationResult {
	if s.dataType.IsReadOnly(op) {
		return s.dataType.ComputeResult(op, s.confirmed)
	} else {
		switch s.role {
		case util.Client:
			for _, c := range s.peers {
				<-c.Send(context.Background(), op)
			}
			var rVal = s.dataType.ComputeResult(op, s.confirmed)
			s.confirmed = append(s.confirmed, op)
			return rVal
		case util.Server:
			var rVal = s.dataType.ComputeResult(op, s.confirmed)
			s.confirmed = append(s.confirmed, op)
			for _, c := range s.peers {
				c.Send(context.Background(), op)
			}
			return rVal
		}
	}

	return nil
}

func (s *Sequencer) Introduce(cid int64, link network.Link) {
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

func (s *Sequencer) Receive(rid int64, msg interface{}) interface{} {
	if op, ok := msg.(Operation); ok {
		s.confirmed = append(s.confirmed, op)

		if s.role == util.Server {
			var recipientList = s.peers
			delete(recipientList, rid)

			for _, c := range s.peers {
				c.Send(context.Background(), op)
			}

			return true
		}
	}

	return nil
}
