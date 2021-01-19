package network

import "context"

type Link interface {
	Send(ctx context.Context, msg interface{}) <-chan interface{}
	AsyncMessage(interface{})
	BlockingMessage(interface{}) interface{}
}

type Peer interface {
	Introduce(rid int64, link Link)
	Receive(src int64, msg interface{}) interface{}
}
