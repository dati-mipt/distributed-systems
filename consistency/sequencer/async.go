package sequencer

import (
	"github.com/dati-mipt/distributed-algorithms/network"
)

type AsyncSequencerServer struct {
	clients map[int64]network.Link
}

func NewAsyncSequencerServer() *AsyncSequencerServer {
	return &AsyncSequencerServer{
		clients: map[int64]network.Link{},
	}
}

func (s *AsyncSequencerServer) Introduce(cid int64, link network.Link) {
	if cid != 0 && link != nil {
		s.clients[cid] = link
	}
}

func (s *AsyncSequencerServer) Receive(rid int64, msg interface{}) interface{} {
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
	server    network.Link
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

func (c *AsyncSequencerClient) Introduce(rid int64, link network.Link) {
	if rid == 0 && link != nil {
		c.server = link
	}
}

func (c *AsyncSequencerClient) Receive(rid int64, msg interface{}) interface{} {
	if op, ok := msg.(Operation); ok {
		c.confirmed = append(c.confirmed, op)
	}
	return nil
}
