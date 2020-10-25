package network

type Peer interface {
	AsyncMessage(interface{})
}

type Responder interface {
	BlockingMessage(interface{}) interface{}
}
