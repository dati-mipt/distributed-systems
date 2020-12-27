package template

import (
	"github.com/dati-mipt/distributed-systems/network"
)

type SequentialServer struct {
	clients map[int64]network.Link
}

func NewSequentialServer() *SequentialServer {
	return &SequentialServer{
		clients: map[int64]network.Link{},
	}
}

func (s *SequentialServer) Introduce(cid int64, link network.Link) {
	if cid != 0 && link != nil {
		s.clients[cid] = link
	}
}

func (s *SequentialServer) Receive(rid int64, msg interface{}) interface{} {
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
	server    network.Link
}

func (c *SequentialClient) Perform(op Operation) OperationResult {
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

func (c *SequentialClient) Introduce(rid int64, link network.Link) {
	if rid == 0 && link != nil {
		c.server = link
	}
}

func (c *SequentialClient) Receive(rid int64, msg interface{}) interface{} {
	if op, ok := msg.(Operation); ok {
		c.confirmed = append(c.confirmed, op)
	}
	return nil
}
