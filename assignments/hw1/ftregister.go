package ftregister

import (
	"fmt"
	"github.com/dati-mipt/distributed-storage-algorithms/network"
	"github.com/dati-mipt/distributed-storage-algorithms/util"
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
	fmt.Printf("write value: %d\n", value)
	
	r.current.Val = value
	r.current.Ts = util.Timestamp{Number: r.current.Ts.Number + 1, Rid: r.rid}


	return true
}

// compound read = {read + write}

func (r *FaultTolerantRegister) Read() int64 {
	fmt.Printf("rid: %d, ts: %d, read value: %d\n", r.current.Ts.Number, r.rid, r.current.Val)

	// read data from replica and set current TimestampedValue due to max replica id

	var msg interface{}
	if update, ok := msg.(util.TimestampedValue); ok {
		if r.current.Ts.Less(update.Ts) {
			r.current = update
		}
	}

	// update Timestamp

	r.current.Ts = util.Timestamp{Number: r.current.Ts.Number + 1, Rid: r.rid}

	// write value to replicas

	for _, rep := range r.replicas {
		rep.AsyncMessage(r.current)
	}

	return r.current.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}

}

// make a copy of data
func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	if update, ok := msg.(util.TimestampedValue); ok {
		if r.current.Ts.Less(update.Ts) {
			r.current = update
		}
	}
	return nil
}

func (r *FaultTolerantRegister) Update() {
	for _, rep := range r.replicas {
		rep.AsyncMessage(r.current)
	}
}
