package hw1

import (
	"context"
	"errors"
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type FaultTolerantRegister struct {
	rid   int64
	links map[int64]network.Link
	ts    util.TimestampedValue
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		rid:   rid,
		links: map[int64]network.Link{},
	}
}

func (r *FaultTolerantRegister) Broadcast(msg interface{}) (channels map[int64]<-chan interface{}) { // map from uint64 to msg from channel
	channels = make(map[int64]<-chan interface{})
	for key, val := range r.links {
		channels[key] = val.Send(context.Background(), msg)
	}
	return
}

func (r *FaultTolerantRegister) Quorum(channels map[int64]<-chan interface{}, threshold int64) (<-chan interface{}, error) { // map from uint64 to msg from channel

	var oks int64 = 0
	var nOks int64 = 0

	ctx := context.Background()

	condVarMut := new(sync.Mutex)
	condVarMut.Lock()
	cv := sync.NewCond(condVarMut)
	mapGuard := new(sync.Mutex)
	msgMap := make(map[int64]interface{}) // not thread-safe

	for id, link := range channels {
		link := link
		id := id
		go func() {
			defer func() {
				if atomic.LoadInt64(&oks) >= threshold || atomic.LoadInt64(&nOks) >= threshold {
					cv.Signal()
				}
			}()

			select {
			case <-ctx.Done():
				return
			case msg, ok := <-link:
				if ok {
					atomic.AddInt64(&oks, 1)
					mapGuard.Lock()
					msgMap[id] = msg
					mapGuard.Unlock()
				} else {
					atomic.AddInt64(&nOks, 1)
				}
			}

		}()
	}

	cv.Wait()
	ctx.Done()
	if atomic.LoadInt64(&oks) >= threshold {
		channel := make(chan interface{})
		valueWithMaxTimestamp := util.TimestampedValue{}
		mapGuard.Lock()
		for _, timestamp := range msgMap {
			if valueWithMaxTimestamp.Ts.Less(timestamp.(util.TimestampedValue).Ts) {
				valueWithMaxTimestamp = timestamp.(util.TimestampedValue)
			}
		}
		mapGuard.Unlock()
		go func() {
			channel <- valueWithMaxTimestamp
		}()

		return channel, nil
	} else {
		return nil, errors.New("network problems")
	}
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	timestamp := r.LocalRead().Ts
	timestamp.Number += 1
	timestamp.Rid = r.rid
	r.LocalWrite(value, timestamp)
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	stampedValue := r.LocalRead()
	r.LocalWrite(stampedValue.Val, stampedValue.Ts)
	return stampedValue.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.links[rid] = link
	}
}

func (r *FaultTolerantRegister) WriteToRegister(value int64, timestamp util.Timestamp) {
	if r.ts.Ts.Less(timestamp) {
		r.ts.Ts = timestamp
		r.ts.Val = value
	}
}

func (r *FaultTolerantRegister) Receive(_ int64, msg interface{}) interface{} {
	decoded := msg.(string)
	cmd := strings.Split(decoded, " ")
	if cmd[0] == "LocalRead" {
		return r.ts
	} else if cmd[0] == "LocalWrite" {
		var err error
		var value, number, rid int64
		if value, err = strconv.ParseInt(cmd[1], 10, 64); err != nil {
			panic("wrong value")
		}
		if number, err = strconv.ParseInt(cmd[2], 10, 64); err != nil {
			panic("wrong value")
		}
		if rid, err = strconv.ParseInt(cmd[3], 10, 64); err != nil {
			panic("wrong value")
		}

		timestamp := util.Timestamp{Number: number, Rid: rid}

		r.WriteToRegister(value, timestamp)

		//if r.ts.Ts.Less(timestamp) {
		//	r.ts.Ts = timestamp
		//	r.ts.Val = value
		//}

		return r.ts
	} else {
		panic("unknown option")
	}
}

func (r *FaultTolerantRegister) LocalRead() util.TimestampedValue {
	future := r.Broadcast("LocalRead")
	if channel, err := r.Quorum(future, int64(len(r.links)/2+1)); err == nil {
		res := <-channel
		return res.(util.TimestampedValue)
	} else {
		panic(err)
	}
}

func (r *FaultTolerantRegister) LocalWrite(value int64, timestamp util.Timestamp) {
	r.WriteToRegister(value, timestamp)
	msg := []string{"LocalWrite", strconv.FormatInt(value, 10), strconv.FormatInt(timestamp.Number, 10), strconv.FormatInt(timestamp.Rid, 10)}
	future := r.Broadcast(strings.Join(msg, " "))
	if _, err := r.Quorum(future, int64(len(r.links)/2)); err != nil {
		panic(err)
	}
}
