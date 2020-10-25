package sequencer

import (
	"github.com/dati-mipt/distributed-algorithms/network"
)

type SequentialServer struct {
	clients []network.Peer
}

func (s SequentialServer) BlockingMessage(msg interface{}) interface{} {
	if op, ok := msg.(Operation); ok {
		for _, c := range s.clients {
			c.AsyncMessage(op)
		}
	}

	return nil
}

type SequentialClient struct {
	dataType  ReplicatedDataType
	confirmed []Operation
	server    network.Responder
}

func (c SequentialClient) Perform(op Operation) OperationResult {
	if c.dataType.IsReadOnly(op) {
		c.dataType.ComputeResult(op, c.confirmed)
	} else {
		c.server.BlockingMessage(op)
		var rVal = c.dataType.ComputeResult(op, c.confirmed)
		c.confirmed = append(c.confirmed, op)
		return rVal
	}

	return nil
}

func (c SequentialClient) AsyncMessage(msg interface{}) {
	if op, ok := msg.(Operation); ok {
		c.confirmed = append(c.confirmed, op)
	}
}
