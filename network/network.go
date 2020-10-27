package network

type Link interface {
	AsyncMessage(interface{})
	BlockingMessage(interface{}) interface{}
}

type Peer interface {
	Introduce(rid int64, link Link)
	Receive(src int64, msg interface{}) interface{}
}
