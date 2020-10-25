package register

import (
	"github.com/dati-mipt/distributed-algorithms/network"
)

type SingleCopyRegisterServer struct {
	current int64
}

func (s *SingleCopyRegisterServer) BlockingMessage(msg interface{}) interface{} {
	if msg == nil {
		return s.current
	}

	if value, ok := msg.(int64); ok {
		s.current = value
		return true
	}

	return nil
}

type SingleCopyRegisterClient struct {
	s network.Responder
}

func (c SingleCopyRegisterClient) Write(value int64) bool {
	if msg, ok := c.s.BlockingMessage(value).(bool); ok {
		return msg
	}

	return false
}

func (c SingleCopyRegisterClient) Read() int64 {
	if msg, ok := c.s.BlockingMessage(nil).(int64); ok {
		return msg
	}

	return 0
}
