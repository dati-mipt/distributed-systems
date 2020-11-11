package network

import (
	"sync"
)

type ReliableLink struct {
	n   *ReliableNetwork
	src int64
	dst int64
}

func newReliableLink(n *ReliableNetwork, src, dst int64) *ReliableLink {
	return &ReliableLink{
		n:   n,
		src: src,
		dst: dst,
	}
}

func (c *ReliableLink) AsyncMessage(msg interface{}) {
	c.n.Send(c.src, c.dst, msg)
}

func (c *ReliableLink) BlockingMessage(msg interface{}) interface{} {
	return <-c.n.Send(c.src, c.dst, msg)
}

type Message struct {
	src  int64
	dst  int64
	data interface{}
	resp chan<- interface{}
}

type ReliableNetwork struct {
	peers    map[int64]Peer
	messages chan Message
	wg       sync.WaitGroup
}

func NewReliableNetwork() *ReliableNetwork {
	return &ReliableNetwork{
		peers:    map[int64]Peer{},
		messages: make(chan Message, 10),
	}
}

func (n *ReliableNetwork) Register(newPid int64, newPeer Peer) {
	if newPeer == nil {
		return
	}

	if _, ok := n.peers[newPid]; ok {
		return
	}

	for exRid, exPeer := range n.peers {
		var linkToExisting = newReliableLink(n, newPid, exRid)
		newPeer.Introduce(exRid, linkToExisting)

		var linkToNew = newReliableLink(n, exRid, newPid)
		exPeer.Introduce(newPid, linkToNew)
	}

	n.peers[newPid] = newPeer
}

func (n *ReliableNetwork) Route() {
	for {
		select {
		case msg := <-n.messages:
			if peer, ok := n.peers[msg.dst]; ok {
				go func() {
					msg.resp <- peer.Receive(msg.src, msg.data)
					n.wg.Done()
				}()
			}
		}
	}
}

func (n *ReliableNetwork) Send(src, dst int64, msg interface{}) <-chan interface{} {
	resp := make(chan interface{}, 1)

	n.wg.Add(1)
	n.messages <- Message{
		src:  src,
		dst:  dst,
		data: msg,
		resp: resp,
	}

	return resp
}

func (n *ReliableNetwork) Wait() {
	n.wg.Wait()
}
