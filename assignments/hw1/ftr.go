package ftr

import (
	"github.com/dati-mipt/distributed-storage-algorithms/network"
	"github.com/dati-mipt/distributed-storage-algorithms/util"
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
		go func(system_chan chan interface{}, link network.Link, value util.TimestampedValue) {
			if msg, ok := link.BlockingMessage(value).(util.TimestampedValue); ok {
				system_chan <- msg // write to chan to unlock waiting of SingWrite()
			}
		}(system_chan, link, value)
	}

	// wait (system_size + 1) / 2 answers
	for i := 0; i < (system_size + 1) / 2; i++ {
		<-system_chan // "Another one have replied to writing"
	}
	return true
}

func (r *FTRegister) SingleRead() util.TimestampedValue {
	system_size := len(r.replicas)
	system_chan := make(chan interface{}, system_size)

	// truing to write in all replicas
	for _, link := range r.replicas {
		go func(system_chan chan interface{}, link network.Link) {
			if msg, ok := link.BlockingMessage(struct{}{}).(util.TimestampedValue); ok {
				system_chan <- msg // results of read
			}
		}(system_chan, link)
	}

	// wait (system_size + 1) / 2 answers, and write the most actual
	value := r.current
	for i := 0; i < (system_size + 1) / 2; i++ {
	msg := <-system_chan // "Another one have replied to reading"

		msg_value := msg.(util.TimestampedValue)
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
	val.Ts.Rid = r.rid
	r.current.Store(val)
	// Write resultes
	r.SingleWrite(val)
	return true
}

func (r *FTRegister) Receive(rid int64, msg interface{}) interface{} {
	if msg == nil {
		return nil
	}

	switch value := msg.(type) {
		case struct{}: {
			return r.current // Read
		}
		case util.TimestampedValue: {
			r.current.Store(value) // Write
		}
	}

	return nil;
}
