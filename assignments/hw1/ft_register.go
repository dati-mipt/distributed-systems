package hw1

import (
	"context"
	"sync"
	"fmt"
	
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

type FaultTolerantRegister struct {
	rid			int64
	current		util.TimestampedValue
	replicas	map[int64]network.Link
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	if false {
		fmt.Printf("fmt stays on\n")
	}
	
	return &FaultTolerantRegister{
		rid:		rid,
		current:	util.TimestampedValue{
			Ts:		util.Timestamp{
				Rid:	rid,
			},
		},
		replicas:	map[int64]network.Link{},
	}
}

func (r *FaultTolerantRegister) gets(getEntireQuorum bool) (chans []<-chan interface{}, quorum_size, subquorum_size int) {
	quorum_size = len(r.replicas)
	subquorum_size = (quorum_size + 2 - quorum_size % 2) / 2
	
	chans_size := quorum_size
	if getEntireQuorum == false {
		chans_size = subquorum_size
	}
	
	chans = make([]<-chan interface{}, chans_size)
	
	i := 0
	for _, rep := range r.replicas {
		chans[i] = rep.Send(context.Background(), struct{}{})
		i++
		if i == chans_size {
			break
		}
	}
	
	return
}

func (r *FaultTolerantRegister) sets(subquorum_size int) {
	i := 0
	for _, rep := range r.replicas {
		rep.Send(context.Background(), r.current)
		i++
		if i == subquorum_size {
			break
		}
	}
}

func (r *FaultTolerantRegister) find_most_recent(chans []<-chan interface{}) (max util.TimestampedValue) {
	max = r.current
	var mu sync.Mutex
	finished := make(chan bool)
	
	for _, ch := range chans {
		go func(ch <-chan interface{}) {
			if val, ok := (<-ch).(util.TimestampedValue); ok {
				mu.Lock()
				max.Store(val)
				mu.Unlock()
			}
			finished <- true
		}(ch)
	}
	
	for range chans {
		<- finished
	}
	
	return
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	chans, _, k := r.gets(true)
	
	maxVal := r.find_most_recent(chans)
	
	r.current.Val = value
	r.current.Ts.Number = maxVal.Ts.Number + 1
	
	r.sets(k)
	
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	chans, _, k := r.gets(false)
	
	maxVal := r.find_most_recent(chans)
	
	r.current.Val = maxVal.Val
	r.current.Ts.Number = maxVal.Ts.Number
	
	r.sets(k)
	
	return r.current.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}
}

func (r *FaultTolerantRegister) Receive(src int64, msg interface{}) interface{} {
	if val, ok := msg.(util.TimestampedValue); ok {
		if r.current.Ts.Less(val.Ts) {
			r.current.Val = val.Val
			r.current.Ts.Number = val.Ts.Number
		}
		return nil
	}
	
	return r.current
}
