package counter

import (
	"github.com/dati-mipt/distributed-algorithms/network"
)

type BroadcastCounter struct {
	current int64

	replicas []network.Peer
}

func (c *BroadcastCounter) Inc() bool {
	c.current++
	for _, p := range c.replicas {
		p.AsyncMessage(struct{}{})
	}
	return true
}

func (c *BroadcastCounter) Read() int64 {
	return c.current
}

func (c *BroadcastCounter) AsyncMessage(interface{}) {
	c.current++
}
