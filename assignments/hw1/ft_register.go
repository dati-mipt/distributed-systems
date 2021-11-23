package hw1

import (
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
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

func (r *FaultTolerantRegister) Write(value int64) bool {
	r.current.Val = value
	r.current.Ts  = util.Timestamp{Number: r.current.Ts.Number + 1, Rid: r.rid}
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	return r.current.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	print("receive ", rid, "\n")
	if update, ok := msg.(util.TimestampedValue); ok {
		if r.current.Ts.Less(update.Ts) {
			r.current = update
		}
	}
	return nil
}
