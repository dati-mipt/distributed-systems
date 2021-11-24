package hw1

import (
	"context"
	"fmt"
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

type FaultTolerantRegister struct {
	rid      int64
	current  util.TimestampedValue
	replicas []network.Link
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{
		replicas: []network.Link{},
		current: util.TimestampedValue{Ts: util.Timestamp{Rid: rid}},
	}
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	if !r.updateFromQuorum() {
		return false
	}
	r.current.Ts.Number += 1
	r.current.Ts.Rid = r.rid
	r.current.Val = value
	if !r.updateQuorum() {
		return false
	}
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	if !r.updateFromQuorum() {
		return 0
	}
	if !r.updateQuorum() {
		return 0
	}
	return r.current.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	r.replicas = append(r.replicas, link)
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	switch value := msg.(type) {
	case util.TimestampedValue:
		r.current.Store(value)
		return true
	case struct{}:
		return r.current
	}
	return nil
}

func (r *FaultTolerantRegister) quorumCount() int {
	return (len(r.replicas) + 1) / 2 + 1
}

// update current value in node r with values in quorum
func (r *FaultTolerantRegister) updateFromQuorum() bool {
	return r.useQuorum(
		struct{}{},
		func(msg interface{}) bool {
			if msg, ok := (msg).(util.TimestampedValue); ok {
				r.current.Store(msg)
				return true
			}
			return false
		})
}

// update values in quorum nodes with current node value
func (r *FaultTolerantRegister) updateQuorum() bool {
	return r.useQuorum(
		r.current,
		func(msg interface{}) bool {
			if _, ok := (msg).(bool); ok {
				return true
			}
			return false
		})
}

// send msg to quorum and call callback on responses
func (r *FaultTolerantRegister) useQuorum(msg interface{}, callback func(msg interface{}) bool) bool {
	successCnt := 1
	curLink := 0

	responses := make(chan interface{})

	updateFunc := func(linkId int) {
		responses<-<-r.replicas[linkId].Send(context.Background(), msg)
	}

	for ; curLink < r.quorumCount() - 1; curLink++ {
		go updateFunc(curLink)
	}

	for successCnt < r.quorumCount() && curLink < len(r.replicas) {
		if callback(<-responses) {
			successCnt++
		} else {
			go updateFunc(curLink);
			curLink++
		}
	}

	for i := 0; i < r.quorumCount() - successCnt; i++ {
		if callback(<-responses) {
			successCnt++
		}
	}

	return successCnt >= r.quorumCount()
}

func (r *FaultTolerantRegister) print() (s string) {
	return fmt.Sprintf("%d %d %d", r.rid, r.current.Ts.Number, r.current.Val)
}

