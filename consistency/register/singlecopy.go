package register

import (
	"github.com/dati-mipt/distributed-algorithms/network"
)

type SingleCopyRegisterServer struct {
	current int64
}

func (s *SingleCopyRegisterServer) ReceiveMessage(rid int64, msg interface{}) interface{} {
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
	server network.Responder
}

func (c SingleCopyRegisterClient) Write(value int64) bool {
	if msg, ok := c.server.BlockingMessage(value).(bool); ok {
		return msg
	}

	return false
}

func (c SingleCopyRegisterClient) Read() int64 {
	if msg, ok := c.server.BlockingMessage(nil).(int64); ok {
		return msg
	}

	return 0
}
