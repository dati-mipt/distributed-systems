package sequencer

import (
	"github.com/dati-mipt/distributed-algorithms/network"
)

type AsyncSequencerServer struct {
	clients []network.Peer
}

func (s *AsyncSequencerServer) ReceiveMessage(rid int64, msg interface{}) interface{} {
	if op, ok := msg.(Operation); ok {
		for _, c := range s.clients {
			c.AsyncMessage(op)
		}
	}

	return nil
}

type AsyncSequencerClient struct {
	dataType  ReplicatedDataType
	confirmed []Operation
	server    network.Peer
}

func (c *AsyncSequencerClient) Perform(op Operation) OperationResult {
	if c.dataType.IsReadOnly(op) {
		return c.dataType.ComputeResult(op, c.confirmed)
	} else if c.dataType.IsUpdateOnly(op) {
		c.server.AsyncMessage(op)
		return true
	}

	return nil
}

func (c *AsyncSequencerClient) ReceiveMessage(rid int64, msg interface{}) interface{} {
	if op, ok := msg.(Operation); ok {
		c.confirmed = append(c.confirmed, op)
	}
	return nil
}
