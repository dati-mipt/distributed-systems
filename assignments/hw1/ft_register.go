package hw1

import (
	"github.com/dati-mipt/distributed-systems/util"
	"github.com/dati-mipt/distributed-systems/network"
)

type FaultTolerantRegister struct {
	rid      int64
	current  util.TimestampedValue
	replicas map[int64]network.Link
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid:      rid,
		replicas: map[int64]network.Link{},
	}
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	cur := r.read()

	cur.Ts.Rid = r.rid
	cur.Ts.Number++
	cur.Val = value

	r.current.Store(cur)

	r.write(cur)
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	cur := r.read()
	r.write(cur)
	return cur.Val
}

func (r *FaultTolerantRegister) write(cur util.TimestampedValue) {
	for _, peer := range r.replicas {
		if peer == r.replicas[r.rid] {
			continue
		}
		peer.AsyncMessage(cur)
	}
}

func (r *FaultTolerantRegister) read() util.TimestampedValue {
	peersNum := len(r.replicas)
	resps := make(chan interface{}, peersNum)
	for _, peer := range r.replicas {
		if peer == r.replicas[r.rid] {
			continue
		}
		go func(peer network.Link) {
			if msg, ok := peer.BlockingMessage(0).(util.TimestampedValue); ok {
				resps <- msg
			}
		}(peer)
	}


	cur := r.current

	for i := 0; i < (peersNum+1)/2; i++ {
		tsvalue := (<-resps).(util.TimestampedValue)
		if cur.Ts.Less(tsvalue.Ts) {
			cur = tsvalue
		}
	}

	return cur
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	if msg == nil {
		return nil
	}

	switch value := msg.(type) {
	case util.TimestampedValue:
		r.current.Store(value)
	default:
		return r.current
	}

	return nil
}
