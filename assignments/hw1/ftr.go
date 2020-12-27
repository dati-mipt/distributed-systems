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

func (r *FTRegister) SendToReplicas(value interface{}) chan interface{} {
	system_size := len(r.replicas)
	system_chan := make(chan interface{}, system_size)

	for _, link := range r.replicas {
		go func(link network.Link) {
			if msg, ok := link.BlockingMessage(value).(util.TimestampedValue); ok {
				system_chan <- msg // results of read
			}
		}(link)
	}
	return system_chan
}

func (r *FTRegister) WaitAnswer(system_chan chan interface{}, value interface{}) interface{} {
	system_size := len(r.replicas)
	for i := 0; i < (system_size + 1) / 2; i++ {
		fmt.Printf("Waiting...\n") //
		msg := <-system_chan // "Another one have replied to reading"
		fmt.Printf("It was %d'th answer\n", i + 1) //

		if(value != nil) { // value == nil if you wait answers in write
			msg_value := msg.(util.TimestampedValue)
			if value.(util.TimestampedValue).Ts.Less(msg_value.Ts) {
				value = msg_value
			}
		}
	}
	return value
}

func (r *FTRegister) SingleWrite(value util.TimestampedValue) bool {
	system_chan := r.SendToReplicas(value);
	fmt.Printf("Wait for %d answers\n", len(r.replicas)) //
	r.WaitAnswer(system_chan, nil)

	return true
}

func (r *FTRegister) SingleRead() util.TimestampedValue {
	system_chan := r.SendToReplicas(struct{}{});
	value := r.current
	fmt.Printf("Wait for %d answers\n", len(r.replicas)) //
	value = r.WaitAnswer(system_chan, value).(util.TimestampedValue)
	return value
}

func (r *FTRegister) Read() int64 {
	fmt.Printf("SingleRead: [%d]\n", r.rid) //
	value := r.SingleRead()
	fmt.Printf("SingleWrite: [%d]\n", r.rid) //
	r.SingleWrite(value)
	return value.Val
}

func (r *FTRegister) Write(value int64) bool {
	// Reading old data
	fmt.Printf("SingleRead: [%d]\n", r.rid) //
	val := r.SingleRead() // return max TimestapedValue
	// Update own data
	val.Val = value
	val.Ts.Number++
	val.Ts.Rid = r.rid
	r.current.Store(val)
	// Write resultes
	fmt.Printf("SingleWrite: [%d]\n", r.rid) //
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
			return r.current
		}
	}

	return nil;
}
