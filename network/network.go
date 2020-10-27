package network

type Peer interface {
	AsyncMessage(interface{})
}

type Responder interface {
	BlockingMessage(interface{}) interface{}
}

type Receiver interface {
	ReceiveMessage(dst int64, msg interface{}) interface{}
}
