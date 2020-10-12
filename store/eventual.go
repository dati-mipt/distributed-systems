package store

import (
	"github.com/dati-mipt/consistency-algorithms/util"
)

type eventualStoreUpdate struct {
	ts    util.Timestamp
	key   int64
	value int64
}

type EventualStore struct {
	rid        int64
	localClock int64

	store map[int64]timedRow

	replicas []util.Receiver
}

func (s EventualStore) Write(key int64, value int64) bool {
	s.localClock++

	var ts = util.Timestamp{
		Number: s.localClock,
		Rid:    s.rid,
	}

	s.store[key] = timedRow{val: value, ts: ts}

	for _, r := range s.replicas {
		r.Message(eventualStoreUpdate{
			ts:    ts,
			key:   key,
			value: value,
		})
	}

	return true
}

func (s EventualStore) Read(key int64) int64 {
	if row, ok := s.store[key]; ok {
		return row.val
	}

	return 0
}

func (s EventualStore) Message(msg interface{}) {
	if cast, ok := msg.(eventualStoreUpdate); ok {
		s.update(cast)
	}
}

func (s EventualStore) update(u eventualStoreUpdate) {
	if row, ok := s.store[u.key]; !ok || row.ts.Less(u.ts) {
		s.store[u.key] = timedRow{val: u.value, ts: u.ts}
	}

	if s.localClock < u.ts.Number {
		s.localClock = u.ts.Number
	}
}
