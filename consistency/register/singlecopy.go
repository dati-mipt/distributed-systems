package register

import (
	"github.com/dati-mipt/distributed-storage-algorithms/network"
)

type SingleCopyRegisterServer struct {
	current int64
}

func (s *SingleCopyRegisterServer) Introduce(rid int64, link network.Link) {}

func (s *SingleCopyRegisterServer) Receive(rid int64, msg interface{}) interface{} {
	if msg == nil {
		return nil
	}

	switch value := msg.(type) {
	case int64:
		s.current = value
		return true
	case struct{}:
		return s.current
	}

	return nil
}

type SingleCopyRegisterClient struct {
	server network.Link
}

func (c *SingleCopyRegisterClient) Write(value int64) bool {
	if msg, ok := c.server.BlockingMessage(value).(bool); ok {
		return msg
	}

	return false
}

func (c *SingleCopyRegisterClient) Read() int64 {
	if msg, ok := c.server.BlockingMessage(struct{}{}).(int64); ok {
		return msg
	}

	return 0
}

func (c *SingleCopyRegisterClient) Introduce(rid int64, link network.Link) {
	if rid == 0 && link != nil {
		c.server = link
	}
}

func (c *SingleCopyRegisterClient) Receive(rid int64, msg interface{}) interface{} {
	return nil
}
