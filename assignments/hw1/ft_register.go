package hw1

import (
	"context"
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

type FaultTolerantRegister struct {
	rid      int64
	current  util.TimestampedValue
	replicas map[int64]network.Link
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		replicas: map[int64]network.Link{},
		current: util.TimestampedValue{Ts: util.Timestamp{Rid: rid}},
	}
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	if !r.updateFromQuorum() {
		return false
	}
	r.current.Ts.Number += 1
	r.current.Ts.Rid = r.rid
	r.current.Val = value
	if !r.updateQuorum() {
		return false
	}
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	if !r.updateFromQuorum() {
		return 0
	}
	if !r.updateQuorum() {
		return 0
	}
	return r.current.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	r.replicas[rid] = link
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	switch value := msg.(type) {
	case util.TimestampedValue:
		r.current.Store(value)
		return true
	case struct{}:
		return r.current
	}
	return nil
}

func (r *FaultTolerantRegister) quorumCount() int {
	return len(r.replicas) / 2 + 1
}

// update current value in node r with values in quorum
func (r *FaultTolerantRegister) updateFromQuorum() bool {
	successCnt := 1
	for _, link := range r.replicas {
		if msg, ok := (<-link.Send(context.Background(), struct{}{})).(util.TimestampedValue); ok {
			r.current.Store(msg)
			successCnt++
			if successCnt >= r.quorumCount() {
				break
			}
		}
	}
	return successCnt >= r.quorumCount()
}

// update values in quorum nodes with current node value
func (r *FaultTolerantRegister) updateQuorum() bool {
	successCnt := 1
	for _, link := range r.replicas {
		if _, ok := (<-link.Send(context.Background(), r.current)).(bool); ok {
			successCnt++
			if successCnt >= r.quorumCount() {
				break
			}
		}
	}
	return successCnt >= r.quorumCount()
}
