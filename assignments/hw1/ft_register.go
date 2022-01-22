package hw1

import (
	"fmt"
	"context"
	"time"
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

type FaultTolerantRegister struct {
	current  util.TimestampedValue
	rid      int64
	replicas map[int64]network.Link
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid:      rid,
		replicas: map[int64]network.Link{},
	}
}

func CollectResponse(ctx context.Context, big_boi chan util.TimestampedValue, 
	input_chan <- chan interface{}) {

	select{
		case <-ctx.Done():
			return 
		case msg := (<-input_chan):
			if Ts, ok := msg.(util.TimestampedValue); ok {
				big_boi <- Ts	
			}

		// default:
		// 	fmt.Println("waitintg")
	}
}

func CountResponses(ctx context.Context, big_boi chan util.TimestampedValue, 
	r *FaultTolerantRegister, i_chan chan int64) {

	i := int64(0)
	for i < int64(len(r.replicas)) {
		select {
			case <-ctx.Done():
				break
			case msg := <-big_boi:
				if r.current.Ts.Less(msg.Ts) {
					r.current.Store(msg)
				}
				i++
		}
	}
	// print("counted")
	i_chan <- i
}

func (r *FaultTolerantRegister) ReadQuorum() bool {
	// first send to every replica
	// and collect responses to the global chan
	big_boi_chan := make(chan util.TimestampedValue)
	chans := make(map[int64] <-chan interface{})

	for i, l := range r.replicas {
		ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(1 * time.Second))
		chans[i] = l.Send(context.Background(), struct{}{})
		go CollectResponse(ctx, big_boi_chan, chans[i]) 
	}

	i_chan := make(chan int64)
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(1 * time.Second))
	go CountResponses(ctx, big_boi_chan, r, i_chan)
	i := <- i_chan
	if i > int64(len(r.replicas)/2) {
		return true
	}
	fmt.Println(i)
	// fmt.Println(len(r.replicas)/2)
	return false
}

func (r *FaultTolerantRegister) WriteQuorum() bool {
	big_boi_chan := make(chan util.TimestampedValue)
	chans := make(map[int64] <-chan interface{})

	for i, l := range r.replicas {
		ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(1 * time.Second))
		chans[i] = l.Send(context.Background(), r.current)
		go CollectResponse(ctx, big_boi_chan, chans[i]) 
	}

	i_chan := make(chan int64)
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(1 * time.Second))
	go CountResponses(ctx, big_boi_chan, r, i_chan)
	i := <- i_chan
	if i > int64(len(r.replicas)/2) {
		// fmt.Println(i)
		return true
	}
	fmt.Println(i)
	return false
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	if !r.ReadQuorum() {
		return false
	}
	r.current.Val = value
	r.current.Ts = util.Timestamp{Number: r.current.Ts.Number + 1, Rid: r.rid}
	if !r.WriteQuorum() {
		return false
	}
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	r.ReadQuorum();
	r.WriteQuorum();
	return r.current.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	switch t := msg.(type) {
	case util.TimestampedValue:
		{
			prev_current := r.current
			if t.Ts.Less(prev_current.Ts) {
				return r.current
			}
			r.current.Store(t)
			return r.current
		}
	case struct{}:
		return r.current
	}
	return nil
}
