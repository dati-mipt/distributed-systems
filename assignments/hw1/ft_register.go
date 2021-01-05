package hw1

import (
	"context"
	"errors"
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
	"sync"
)

type FaultTolerantRegister struct {
	rid      int64
	state    util.TimestampedValue
	mutex    sync.RWMutex
	replicas map[int64]network.Link
}

const (
	FTR_DISCOVER int32 = iota
	FTR_PROPAGATE
	FTR_STATE
	FTR_ACCEPTED
)

type FTRMessage struct {
	req   int32
	state util.TimestampedValue
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid:      rid,
		state:    util.TimestampedValue{
			Val: 0,
			Ts: util.Timestamp {
				Number: 0,
				Rid:    rid,
			},
		},
		mutex:    sync.RWMutex{},
		replicas: map[int64]network.Link{},
	}
}

func (r *FaultTolerantRegister) updState(val util.TimestampedValue) util.TimestampedValue {
	if cur := r.getState(); !(cur.Ts.Less(val.Ts)) {
		return cur
	}
	r.mutex.Lock()
	if r.state.Ts.Less(val.Ts) {
		r.state = val
	} else {
		val = r.state
	}
	r.mutex.Unlock()
	return val
}

func (r *FaultTolerantRegister) getState() util.TimestampedValue {
	r.mutex.RLock()
	val := r.state
	r.mutex.RUnlock()
	return val
}

func (r *FaultTolerantRegister) Broadcast(ctx context.Context, msgchan chan FTRMessage, msg FTRMessage) {
	var wg sync.WaitGroup
	for _, rep := range r.replicas {
		wg.Add(1)
		go func(dst network.Link) {
			replyCh := dst.Send(ctx, msg)
			if reply, ok := <-replyCh; ok {
				msgchan <-reply.(FTRMessage)
			}
			wg.Done()
		}(rep)
	}
	wg.Wait()
	close(msgchan)
}

func (r *FaultTolerantRegister) Discover() (util.TimestampedValue, bool) {
	newest := r.getState()
	msgchan := make(chan FTRMessage)
	ctx, cancel := context.WithCancel(context.Background())

	go r.Broadcast(ctx, msgchan, FTRMessage{FTR_DISCOVER, newest})

	var qsz int64 = int64(len(r.replicas))/2 + 1
	var count int64 = 1

	for msg, ok := <-msgchan; ok && (count < qsz); count++ {
		if newest.Ts.Less(msg.state.Ts) {
			newest = msg.state
		}
	}
	cancel()
	for _ = range msgchan { }

	return newest, count >= qsz
}

func (r *FaultTolerantRegister) Propagate(newest util.TimestampedValue) bool {
	r.updState(newest)
	msgchan := make(chan FTRMessage)
	ctx, cancel := context.WithCancel(context.Background())

	go r.Broadcast(ctx, msgchan, FTRMessage{FTR_PROPAGATE, newest})

	var qsz int64 = int64(len(r.replicas))/2 + 1
	var count int64 = 1

	for _, ok := <-msgchan; ok && (count < qsz); count++ {
		// nop
	}
	cancel()
	for _ = range msgchan { }

	return count >= qsz
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	newest, ok := r.Discover()
	if !ok {
		return false
	}
	newest.Ts = util.Timestamp{Number: newest.Ts.Number + 1, Rid: r.rid}
	newest.Val = value

	return r.Propagate(newest)
}

func (r *FaultTolerantRegister) Read() int64 {
	newest, ok := r.Discover()
	if !ok {
		panic(errors.New("Discover failed"))
	}
	if !r.Propagate(newest) {
		panic(errors.New("Propagate failed"))
	}
	return newest.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	typed := msg.(FTRMessage)

	switch typed.req {
	case FTR_DISCOVER:
		return FTRMessage{FTR_STATE, r.getState()}
	case FTR_PROPAGATE:
		return FTRMessage{FTR_ACCEPTED, r.updState(typed.state)} // QUEUE?
	default:
		panic(errors.New("Bad request"))
	}
	return nil
}
