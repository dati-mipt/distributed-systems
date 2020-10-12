package store

import (
	"github.com/dati-mipt/consistency-algorithms/util"
)

type eventualStoreUpdate struct {
	key   int64
	value util.TimestampedValue
}

type EventualStore struct {
	rid        int64
	localClock int64

	store map[int64]util.TimestampedValue

	replicas []util.Receiver
}

func (s EventualStore) Write(key int64, value int64) bool {
	s.localClock++

	var tValue = util.TimestampedValue{
		Val: value,
		Ts: util.Timestamp{
			Number: s.localClock,
			Rid:    s.rid,
		},
	}

	s.store[key] = tValue

	for _, r := range s.replicas {
		r.Message(eventualStoreUpdate{
			key:   key,
			value: tValue,
		})
	}

	return true
}

func (s EventualStore) Read(key int64) int64 {
	if row, ok := s.store[key]; ok {
		return row.Val
	}

	return 0
}

func (s EventualStore) Message(msg interface{}) {
	if cast, ok := msg.(eventualStoreUpdate); ok {
		s.update(cast)
	}
}

func (s EventualStore) update(u eventualStoreUpdate) {
	if row, ok := s.store[u.key]; !ok || row.Ts.Less(u.value.Ts) {
		s.store[u.key] = u.value
	}

	if s.localClock < u.value.Ts.Number {
		s.localClock = u.value.Ts.Number
	}
}
