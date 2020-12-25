package ftr

import (
	"github.com/dati-mipt/distributed-storage-algorithms/network"
	"github.com/dati-mipt/distributed-storage-algorithms/util"
	"fmt"
)

type FTRegister struct {
	rid      int64
	current  util.TimestampedValue

	replicas map[int64]network.Link
}

func NewFTRegister(rid int64) *FTRegister {
	return &FTRegister{
		rid:      rid,
		replicas: map[int64]network.Link{},
	}
}

func (r *FTRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}
}

func (r *FTRegister) SingleWrite(value util.TimestampedValue) bool {
	system_size := len(r.replicas)
	system_chan := make(chan interface{}, system_size)

	// truing to write in all replicas
	for _, link := range r.replicas {
		go func(link neework.Link, value util.TimestampedValue) {
			if msg, ok := link.BlockingMessage(value).(util.TimestampedValue); ok {
				system_chan <- msg // write to chan to unlock waiting of SingWrite()
			}
		}(link, value)
	}

	// wait (system_size + 1) / 2 answers
	for i := 0; i < (system_size + 1) / 2; ++i {
		msg <- system_chan // "Another one have replied to writing"
	}
	return true
}

func (r *FTRegister) SingleRead() util.TimestapedValue {
	system_size := len(r.replicas)
	system_chan := make(chan interface{}, system_size)

	// truing to write in all replicas
	for _, link := range r.replicas {
		go func(link neework.Link, value util.TimestampedValue) {
			if msg, ok := link.BlockingMessage(value).(util.TimestampedValue); ok {
				system_chan <- msg // write to chan the answer
			}
		}(link, struct{}{})
	}

	// wait (system_size + 1) / 2 answers, and write the most actual
	value := r.current
	for i := 0; i < (system_size + 1) / 2; ++i {
		msg <- system_chan // "Another one have replied to reading"

		msg_value := msg.(util.TimestapedValue)
		if value.Ts.Less(msg_value.Ts) {
			value = msg_value
		}
	}
	return value
}

func (r *FTRegister) Read() int64 {
	value := r.SingleRead()
	r.SingleWrite(value)
	return value.Val
}

func (r *FTRegister) Write(value int64) bool {
	// Reading old data
	val := r.SingleRead() // return max TimestapedValue
	// Update own data
	val.Val = value
	val.Ts.Number++
	val.Ts.Rid = rid
	r.current.Store(val)
	// Write resultes
	r.SingleWrite(val)
	return true
}

func (r *FTRegister) Receive(rid int64, msg interface{}) interface{} {
	if msg == nil {
		return nil
	}

	value := msg.(type)
	if value == struct{} {
		return r.current // Read
	} else if value == util.TimestampedValue {
		r.current.Store(value) // Write
	}

	return nil
}
