package store

import "github.com/dati-mipt/consistency-algorithms/util"

type storeReplica interface {
	Update(key int64, value int64, ts util.Timestamp)
}

type timedRow struct {
	ts  util.Timestamp
	val int64
}
type EventualStore struct {
	rid        int64
	localClock int64

	store map[int64]timedRow

	replicas []storeReplica
}

func (s EventualStore) Write(key int64, value int64) bool {
	s.localClock++

	var ts = util.Timestamp{
		Number: s.localClock,
		Rid:    s.rid,
	}

	s.store[key] = timedRow{val: value, ts: ts}

	for _, r := range s.replicas {
		r.Update(key, value, ts)
	}

	return true
}
func (s EventualStore) Read(key int64) int64 {
	if row, ok := s.store[key]; ok {
		return row.val
	}

	return 0
}
func (s EventualStore) Update(key int64, value int64, ts util.Timestamp) {
	if row, ok := s.store[key]; !ok || row.ts.Less(ts) {
		s.store[key] = timedRow{val: value, ts: ts}
	}

	if s.localClock < ts.Number {
		s.localClock = ts.Number
	}
}
