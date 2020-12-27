package register

import (
	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

type EpidemicRegister struct {
	rid      int64
	current  util.TimestampedValue
	replicas map[int64]network.Link
}

func NewEpidemicRegister(rid int64) *EpidemicRegister {
	return &EpidemicRegister{
		rid:      rid,
		replicas: map[int64]network.Link{},
	}
}

func (r *EpidemicRegister) Write(value int64) bool {
	r.current.Val = value
	r.current.Ts = util.Timestamp{Number: r.current.Ts.Number + 1, Rid: r.rid}
	return true
}

func (r *EpidemicRegister) Read() int64 {
	return r.current.Val
}

func (r *EpidemicRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}
}

func (r *EpidemicRegister) Receive(rid int64, msg interface{}) interface{} {
	if update, ok := msg.(util.TimestampedValue); ok {
		if r.current.Ts.Less(update.Ts) {
			r.current = update
		}
	}
	return nil
}

func (r *EpidemicRegister) Periodically() {
	for _, rep := range r.replicas {
		rep.AsyncMessage(r.current)
	}
}
