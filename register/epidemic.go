package register

import "github.com/dati-mipt/consistency-algorithms/util"

type LatestValueReplica interface {
	Latest(value int64, ts util.Timestamp)
}

type EpidemicRegister struct {
	rid int64

	current int64
	written util.Timestamp

	replicas []LatestValueReplica
}

func (r EpidemicRegister) Write(value int64) bool {
	r.current = value
	r.written = util.Timestamp{Number: r.written.Number + 1, Rid: r.rid}
	return true
}

func (r EpidemicRegister) Read() int64 {
	return r.current
}

func (r EpidemicRegister) Periodically() {
	for _, rep := range r.replicas {
		rep.Latest(r.current, r.written)
	}
}

func (r EpidemicRegister) Latest(value int64, ts util.Timestamp) {
	if r.written.Less(ts) {
		r.current = value
		r.written = ts
	}
}
