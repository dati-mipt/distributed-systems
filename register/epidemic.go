package register

import "github.com/dati-mipt/consistency-algorithms/util"

type timestampedValue struct {
	value int64
	ts    util.Timestamp
}

type EpidemicRegister struct {
	rid int64

	current int64
	written util.Timestamp

	replicas []util.Receiver
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
		rep.Message(timestampedValue{r.current, r.written})
	}
}

func (r EpidemicRegister) Message(msg interface{}) {
	if cast, ok := msg.(timestampedValue); ok {
		r.update(cast)
	}
}

func (r EpidemicRegister) update(u timestampedValue) {
	if r.written.Less(u.ts) {
		r.current = u.value
		r.written = u.ts
	}
}
