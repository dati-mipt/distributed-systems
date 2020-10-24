package sequencer

import "github.com/dati-mipt/distributed-algorithms/util"

type AsyncServer struct {
	clients []util.Peer
}

func (s AsyncServer) Message(msg interface{}) {
	if op, ok := msg.(Operation); ok {
		for _, c := range s.clients {
			c.Message(op)
		}
	}
}

type AsyncClient struct {
	dataType  ReplicatedDataType
	confirmed []Operation
	server    util.Peer
}

func (c AsyncClient) Perform(op Operation) OperationResult {
	if c.dataType.IsReadOnly(op) {
		return c.dataType.ComputeResult(op, c.confirmed)
	} else if c.dataType.IsUpdateOnly(op) {
		c.server.Message(op)
		return true
	}

	return nil
}

func (c AsyncClient) Message(msg interface{}) {
	if op, ok := msg.(Operation); ok {
		c.confirmed = append(c.confirmed, op)
	}
}
