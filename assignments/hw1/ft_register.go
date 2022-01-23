package hw1

import (
	"context"
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
		rid: rid,
		current: util.TimestampedValue{
			Ts: util.Timestamp{
				Rid: rid,
			},
		},
		replicas: []network.Link{},
	}
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	if (!r.NodeValueUpdateFromQuorum()) {
		return false
	}
	r.current.Ts.Rid = r.rid
	r.current.Val = value
	r.current.Ts.Number++
	if (!r.QuorumUpdateValue()) {
		return false
	}
	return true
}

func (r *FaultTolerantRegister) Read() int64 {
	if (!r.NodeValueUpdateFromQuorum()) {
		return 0
	}
	if (!r.QuorumUpdateValue()) {
		return 0
	}
  
	return r.current.Val
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
	r.replicas = append(r.replicas, link)
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
  switch value := msg.(type) {
	case struct{}:
		return r.current
  case util.TimestampedValue:
		r.current.Store(value)
		return true
	}
	return nil
}

func (r *FaultTolerantRegister) QuorumUpdateValue() bool {
	responses := make(chan interface{})
	numLink := r.SendMsgToQuorum(r.current, responses)

	return r.HandleResponses(
		numLink, responses, r.current,
		func(msg interface{}) bool {
			_, ok := msg.(bool)
			return ok
		},
	)
}

func (r *FaultTolerantRegister) NodeValueUpdateFromQuorum() bool {
	responses := make(chan interface{})
	numLink := r.SendMsgToQuorum(r.current, responses)

	return r.HandleResponses(
		numLink, responses, struct{}{},
		func(msg interface{}) bool {
			message, ok := msg.(util.TimestampedValue)
			if (ok) {
				r.current.Store(message)
			}
			return ok
		},
	)
}

func (r *FaultTolerantRegister) HandleResponses(numLink int, responses chan interface{}, msg interface{}, handle func(msg interface{}) bool) bool {
	numSuccess := 1
	for numLink < len(r.replicas) && numSuccess < r.QuorumSize() {
		if (handle(<-responses)) {
			numSuccess++
		} else {
			go r.SendMsg(responses, msg, numLink)
			numLink++
		}
	}

	for i := 0; i < r.QuorumSize() - numSuccess; i++ {
		if (handle(<-responses)) {
			numSuccess++
		}
	}

	return numSuccess >= r.QuorumSize()
}

func (r *FaultTolerantRegister) SendMsg(responses chan interface{}, msg interface{}, numLink int) {
	responses <- <-r.replicas[numLink].Send(context.Background(), msg)
}

func (r *FaultTolerantRegister) SendMsgToQuorum(msg interface{}, responses chan interface{}) int {
	currentLink := 0
	for ; currentLink < r.QuorumSize() - 1; currentLink++ {
		go r.SendMsg(responses, msg, currentLink)
	}
  
	return currentLink
}

func (r *FaultTolerantRegister) QuorumSize() int {
	return (len(r.replicas) + 1) / 2
}
