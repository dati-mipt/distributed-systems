package sequencer

import (
	"github.com/dati-mipt/distributed-algorithms/network"
)

type SequentialServer struct {
	clients map[int64]network.Peer
}

func (s SequentialServer) ReceiveMessage(rid int64, msg interface{}) interface{} {
	var recipientList = s.clients
	delete(recipientList, rid)

	for _, c := range recipientList {
		c.AsyncMessage(msg)
	}

	return true
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

func (c SequentialClient) ReceiveMessage(rid int64, msg interface{}) interface{} {
	if op, ok := msg.(Operation); ok {
		c.confirmed = append(c.confirmed, op)
	}
	return nil
}
