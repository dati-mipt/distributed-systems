package hw1

import (
	"context"
	"math"
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

type FaultTolerantRegister struct {
	rid		int64
	links	map[int64]network.Link
	ts		util.TimestampedValue
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid: rid,
		links: map[int64]network.Link{},
	}
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	r.ts.Ts.Number++
	r.ts.Val = value
	if max_ts := r.quorumRead(); r.ts.Ts.Less(max_ts.Ts) {
		r.ts = max_ts
	}
	r.quorumWrite(r.ts)
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	max_ts := r.quorumRead()
	if r.ts.Ts.Less(max_ts.Ts) {
		r.ts = max_ts
	} else {
		r.quorumWrite(r.ts)
	}
	return r.ts.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.links[rid] = link
	}
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	if msg == nil {
		return r.ts
	} else {
		if r.ts.Ts.Less(msg.(util.TimestampedValue).Ts) {
			r.ts = msg.(util.TimestampedValue)
		}
		return nil
	}
}

func (r *FaultTolerantRegister) quorumSend(msg interface{}) (channels map[int64]<-chan interface{}) {
	channels = make(map[int64]<-chan interface{})
	for key, val := range r.links {
		channels[key] = val.Send(context.Background(), msg)
	}
	return
}

func (r *FaultTolerantRegister) quorumRead() util.TimestampedValue {
	channels := r.quorumSend(nil)
	max_ts := util.TimestampedValue{Ts: util.Timestamp{Number: 0}}
	for res_count := 0; res_count < int(math.Ceil((float64(len(r.links) + 1))/2)); {
		for _, val := range channels {
			if res, ok := <-val; ok {
				res_count++
				if max_ts.Ts.Less(res.(util.TimestampedValue).Ts) {
					max_ts = res.(util.TimestampedValue)
				}
			}
		}
	}
	return max_ts
}

func (r *FaultTolerantRegister) quorumWrite(ts util.TimestampedValue) {
	channels := r.quorumSend(r.ts)
	for res_count := 0; res_count < int(math.Ceil((float64(len(r.links) + 1)/2))); {
		for _, val := range channels {
			if _, ok := <-val; ok {
				res_count++
			}
		}
	}
}

