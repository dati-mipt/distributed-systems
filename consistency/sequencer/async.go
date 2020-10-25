package sequencer

import (
	"github.com/dati-mipt/distributed-algorithms/network"
)

type AsyncServer struct {
	clients []network.Peer
}

func (s AsyncServer) AsyncMessage(msg interface{}) {
	if op, ok := msg.(Operation); ok {
		for _, c := range s.clients {
			c.AsyncMessage(op)
		}
	}
}

type AsyncClient struct {
	dataType  ReplicatedDataType
	confirmed []Operation
	server    network.Peer
}

func (c AsyncClient) Perform(op Operation) OperationResult {
	if c.dataType.IsReadOnly(op) {
		return c.dataType.ComputeResult(op, c.confirmed)
	} else if c.dataType.IsUpdateOnly(op) {
		c.server.AsyncMessage(op)
		return true
	}

	return nil
}

func (c AsyncClient) AsyncMessage(msg interface{}) {
	if op, ok := msg.(Operation); ok {
		c.confirmed = append(c.confirmed, op)
	}
}
