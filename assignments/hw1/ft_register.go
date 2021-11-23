package hw1

import (
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
	"context"
)

type FaultTolerantRegister struct {
	current   util.TimestampedValue
	rid       int64
	replicas  map[int64]network.Link
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid:      rid,
		replicas: map[int64]network.Link{},
	}
}

func (r *FaultTolerantRegister) ReadQuorum() bool {
	i := 0
	for _, l := range r.replicas {
		if msg, ok := (<-l.Send(context.Background(), struct{}{})).(util.TimestampedValue); ok{
			r.current.Store(msg)
			i++
		}
	}
	if (i > len(r.replicas) / 2) {
		return true
	}
	return false
}

func (r *FaultTolerantRegister) WriteQuorum() bool {
	i := 0
	for _, l := range r.replicas {
		if _, ok := (<-l.Send(context.Background(), r.current)).(util.TimestampedValue); ok{
			i++
		}
	}
	if (i > len(r.replicas) / 2) {
		return true
	}
	return false
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	if !r.ReadQuorum() {
		return false
	}
	r.current.Val = value
	r.current.Ts  = util.Timestamp{Number: r.current.Ts.Number + 1, Rid: r.rid}
	if !r.WriteQuorum() {
		return false
	}
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	if (r.ReadQuorum() && r.WriteQuorum()) {
		return r.current.Val
	}
	return 404 //?????
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	switch t:= msg.(type) {
	case util.TimestampedValue: {
		r.current.Store(t)
		return r.current
	}
	case  struct{}:
		return r.current
	}
	return nil
}
