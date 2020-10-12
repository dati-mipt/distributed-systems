package counter

import "github.com/dati-mipt/distributed-algorithms/util"

type BroadcastCounter struct {
	current int64

	replicas []util.Receiver
}

func (c BroadcastCounter) Inc() bool {
	c.current++
	for _, p := range c.replicas {
		p.Message(struct{}{})
	}
	return true
}

func (c BroadcastCounter) Read() int64 {
	return c.current
}

func (c BroadcastCounter) Message(interface{}) {
	c.current++
}
