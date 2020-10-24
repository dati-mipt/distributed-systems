package sequencer

import "github.com/dati-mipt/distributed-algorithms/util"

type SequentialServer struct {
	clients []util.Peer
}

func (s SequentialServer) SendAndWaitForResponse(msg interface{}) interface{} {
	if op, ok := msg.(Operation); ok {
		for _, c := range s.clients {
			c.Message(op)
		}
	}

	return nil
}

type SequentialClient struct {
	dataType  ReplicatedDataType
	confirmed []Operation
	server    util.Responder
}

func (c SequentialClient) Perform(op Operation) OperationResult {
	if c.dataType.IsReadOnly(op) {
		c.dataType.ComputeResult(op, c.confirmed)
	} else {
		c.server.SendAndWaitForResponse(op)
		var rVal = c.dataType.ComputeResult(op, c.confirmed)
		c.confirmed = append(c.confirmed, op)
		return rVal
	}

	return nil
}

func (c SequentialClient) Message(msg interface{}) {
	if op, ok := msg.(Operation); ok {
		c.confirmed = append(c.confirmed, op)
	}
}
