package counter

type IncrementalReplica interface {
	IncMessage()
}

type BroadcastCounter struct {
	current int64

	replicas []IncrementalReplica
}

func (c BroadcastCounter) Inc() bool {
	c.current++
	for _, p := range c.replicas {
		p.IncMessage()
	}
	return true
}

func (c BroadcastCounter) Read() int64 {
	return c.current
}

func (c BroadcastCounter) IncMessage() {
	c.current++
}
