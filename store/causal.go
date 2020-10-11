package store

import "github.com/dati-mipt/consistency-algorithms/util"

type causalStoreUpdate struct {
	ts    util.Timestamp
	key   int64
	value int64
	deps  map[int64]util.Timestamp
}

type causalStoreReplica interface {
	update(u causalStoreUpdate)
}

type inBuffer struct {
	updates       []causalStoreUpdate
	lastProcessed int64
}

func (b inBuffer) empty() bool {
	return len(b.updates) == 0
}

func (b inBuffer) last() causalStoreUpdate {
	if b.empty() {
		return causalStoreUpdate{}
	}
	return b.updates[0]
}

func (b inBuffer) enqueue(u causalStoreUpdate) {
	b.updates = append(b.updates, u)
}

func (b inBuffer) dequeue() causalStoreUpdate {
	if b.empty() {
		return causalStoreUpdate{}
	}

	// todo: not concurrent-friendly
	var value = b.updates[0]
	b.updates = b.updates[1:]
	return value
}

type CausalStore struct {
	rid        int64
	localClock int64

	store   map[int64]timedRow
	buffers map[int64]inBuffer
	deps    map[int64]util.Timestamp

	replicas []causalStoreReplica
}

func (s CausalStore) Write(key int64, value int64) bool {
	var updates = s.buffers[s.rid]
	updates.lastProcessed++
	s.buffers[s.rid] = updates

	var ts = util.Timestamp{
		Number: s.buffers[s.rid].lastProcessed,
		Rid:    s.rid,
	}

	s.store[key] = timedRow{val: value, ts: ts}

	for _, r := range s.replicas {
		r.update(causalStoreUpdate{
			ts:    ts,
			key:   key,
			value: value,
			deps:  s.deps,
		})
	}

	return true
}

func (s CausalStore) Read(key int64) int64 {
	if row, ok := s.store[key]; ok {
		return row.val
	}

	return 0
}

func (s CausalStore) update(u causalStoreUpdate) {
	var buffer = s.buffers[u.ts.Rid]
	buffer.enqueue(u)
	s.buffers[u.ts.Rid] = buffer
}

func (s CausalStore) readyToApply(u causalStoreUpdate) bool {
	var ready = true
	for _, ts := range u.deps {
		if s.buffers[ts.Rid].lastProcessed <= ts.Number {
			ready = false
		}
	}

	return ready
}

func (s CausalStore) Periodically() {
	for rid, buffer := range s.buffers {
		if !buffer.empty() && s.readyToApply(buffer.last()) {
			var u = buffer.dequeue()

			if row, ok := s.store[u.key]; !ok || row.ts.Less(u.ts) {
				s.store[u.key] = timedRow{val: u.value, ts: u.ts}
			}

			buffer.lastProcessed = u.ts.Number

			if u.ts.Number > buffer.lastProcessed { //keep up with time
				buffer.lastProcessed = u.ts.Number
			}

			s.buffers[rid] = buffer
		}
	}
}
