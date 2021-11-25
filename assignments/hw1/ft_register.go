package hw1

import (
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
	"math"

	"context"
	//"fmt"
)

type FaultTolerantRegister struct {
	rid     int64
	links   map[int64]network.Link
	current util.TimestampedValue
}

type Msg struct {
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid:     rid,
		links:   map[int64]network.Link{},
		current: util.TimestampedValue{Ts: util.Timestamp{Rid: rid}},
	}
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	r.current.Val = value
	r.current.Ts = util.Timestamp{Number: r.current.Ts.Number + 1, Rid: r.rid}
	r.current = r.readQuorum()
	r.writeQuorum(r.current)
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	ret := r.readQuorum()
	r.current = ret
	return ret.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.links[rid] = link
	}
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	if update, ok := msg.(util.TimestampedValue); ok {
		if r.current.Ts.Less(update.Ts) {
			r.current = update
		}
		return nil
	} else {
		return r.current
	}
}

func (r *FaultTolerantRegister) readQuorum() util.TimestampedValue {
	chans := make(map[int64]<-chan interface{})

	for key, val := range r.links {
		chans[key] = val.Send(context.Background(), nil)
	}

	maxTs := r.current

	for i := 0; i < int(math.Ceil(float64((len(r.links)+1)/2))); {
		for _, val := range chans {
			if ret, ok := <-val; ok {
				if ret != nil {
					i++
					if maxTs.Ts.Less(ret.(util.TimestampedValue).Ts) {
						maxTs = ret.(util.TimestampedValue)
					}
				}
			}
		}
	}
	return maxTs
}

func (r *FaultTolerantRegister) writeQuorum(ts util.TimestampedValue) {
	chans := make(map[int64]<-chan interface{})

	for key, val := range r.links {
		chans[key] = val.Send(context.Background(), ts)
	}

	for i := 0; i < int(math.Ceil(float64((len(r.links)+1)/2))); {
		for _, val := range chans {
			if ret, ok := <-val; ok {
				if ret == nil {
					i++
				}
			}
		}
	}
}
