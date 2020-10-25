package network

type Message struct {
	src  int64
	dst  int64
	data interface{}
}

type ReliableNetwork struct {
	peers    map[int64]Peer
	messages chan Message
}

type ReliableLink struct {
	n   *ReliableNetwork
	src int64
	dst int64
}

func (n *ReliableNetwork) Send(src, dst int64, msg interface{}) <-chan interface{} {
	var mock <-chan interface{}
	return mock
}

func (c *ReliableLink) AsyncMessage(msg interface{}) {
	c.n.Send(c.src, c.dst, msg)
}

func (c *ReliableLink) BlockingMessage(msg interface{}) interface{} {
	return <-c.n.Send(c.src, c.dst, msg)
}
