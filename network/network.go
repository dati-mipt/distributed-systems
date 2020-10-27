package network

type Link interface {
	AsyncMessage(interface{})
	BlockingMessage(interface{}) interface{}
}

type Peer interface {
	Receive(src int64, msg interface{}) interface{}
	Introduce(rid int64, link Link)
}
