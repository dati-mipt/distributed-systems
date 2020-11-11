package register

import (
	//"fmt"
	"github.com/dati-mipt/distributed-storage-algorithms/network"
	"github.com/dati-mipt/distributed-storage-algorithms/util"
	"sync"
)

type FaultTolerantRegister struct {
	rid      int64
	replicas map[int64]network.Link

	mu      sync.Mutex
	current util.TimestampedValue
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid:      rid,
		replicas: map[int64]network.Link{},
	}
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	cur := r.rawRead() // To make logical time monotonic

	cur.Ts.Rid = r.rid
	cur.Ts.Number++
	cur.Val = value
	r.mu.Lock()
	r.current.Store(cur)
	r.mu.Unlock()

	r.rawWrite(cur)
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	cur := r.rawRead()
	r.rawWrite(cur) // To achieve monotonic reads

	return cur.Val
}

func (r *FaultTolerantRegister) rawWrite(cur util.TimestampedValue) {
	for _, peer := range r.replicas {
		if peer == r.replicas[r.rid] {
			continue
		}
		peer.AsyncMessage(cur)
	}
}

func (r *FaultTolerantRegister) rawRead() util.TimestampedValue {
	peers_num := len(r.replicas)
	resps := make(chan interface{}, peers_num)
	for _, peer := range r.replicas {
		if peer == r.replicas[r.rid] {
			continue
		}
		go func() {
			if msg, ok := peer.BlockingMessage(struct{}{}).(util.TimestampedValue); ok {
				resps <- msg
			}
		}()
	}

	// In this design there are no clients,
	// each 'client' is part of the distributed system,
	// therefore it only iterates over (peers_num+1)/2 nodes,
	// but uses its own value
	r.mu.Lock()
	cur := r.current
	r.mu.Unlock()
	for i := 0; i < (peers_num+1)/2; i++ {
		tsvalue := (<-resps).(util.TimestampedValue)
		if cur.Ts.Less(tsvalue.Ts) {
			cur = tsvalue
		}
	}
	return cur
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	if msg == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	switch value := msg.(type) {
	case struct{}:
		return r.current
	case util.TimestampedValue:
		r.current.Store(value)
	}

	return nil
}
