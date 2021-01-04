package register

import (
	"context"
	"github.com/dati-mipt/distributed-systems/network"
)

type SingleCopyRegister struct {
	isServer bool

	current int64
	server  network.Link
}

func (r *SingleCopyRegister) Write(value int64) bool {
	if r.isServer {
		r.current = value
		return true
	}

	if msg, ok := (<-r.server.Send(context.Background(), value)).(bool); ok {
		return msg
	}

	return false
}

func (r *SingleCopyRegister) Read() int64 {
	if r.isServer {
		return r.current
	}

	if msg, ok := (<-r.server.Send(context.Background(), struct{}{})).(int64); ok {
		return msg
	}

	return 0
}

func (r *SingleCopyRegister) Introduce(rid int64, link network.Link) {
	if !r.isServer && link != nil {
		r.server = link
	}
}

func (r *SingleCopyRegister) Receive(rid int64, msg interface{}) interface{} {
	if !r.isServer {
		return nil
	}

	switch value := msg.(type) {
	case int64:
		r.current = value
		return true
	case struct{}:
		return r.current
	}

	return nil
}
