package store

import (
	"context"

	"github.com/dati-mipt/distributed-systems/network"
	"github.com/dati-mipt/distributed-systems/util"
)

// todo: move queue from store module module

type causalStoreUpdate struct {
	key   int64
	value util.TimestampedValue
	deps  map[int64]util.Timestamp
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
	store      map[int64]util.TimestampedValue
	buffers    map[int64]inBuffer
	deps       map[int64]util.Timestamp
	replicas   map[int64]network.Link
}

func NewCausalStore(rid int64) *CausalStore {
	return &CausalStore{
		rid:      rid,
		store:    map[int64]util.TimestampedValue{},
		buffers:  map[int64]inBuffer{},
		deps:     map[int64]util.Timestamp{},
		replicas: map[int64]network.Link{},
	}
}

func (s *CausalStore) Write(key int64, value int64) bool {
	var updates = s.buffers[s.rid]
	updates.lastProcessed++
	s.buffers[s.rid] = updates

	var tValue = util.TimestampedValue{
		Val: value,
		Ts: util.Timestamp{
			Number: s.localClock,
			Rid:    s.rid,
		},
	}

	s.store[key] = tValue

	for _, r := range s.replicas {
		r.Send(context.Background(), causalStoreUpdate{
			key:   key,
			value: tValue,
			deps:  s.deps,
		})
	}

	return true
}

func (s *CausalStore) Read(key int64) int64 {
	if row, ok := s.store[key]; ok {
		return row.Val
	}

	return 0
}

func (s *CausalStore) Introduce(rid int64, link network.Link) {
	if rid != 0 && link != nil {
		s.replicas[rid] = link
	}
}

func (s *CausalStore) Receive(rid int64, msg interface{}) interface{} {
	if update, ok := msg.(causalStoreUpdate); ok {
		var buffer = s.buffers[update.value.Ts.Rid]
		buffer.enqueue(update)
		s.buffers[update.value.Ts.Rid] = buffer
	}
	return nil
}

func (s *CausalStore) readyToApply(u causalStoreUpdate) bool {
	var ready = true
	for _, ts := range u.deps {
		if s.buffers[ts.Rid].lastProcessed <= ts.Number {
			ready = false
		}
	}

	return ready
}

func (s *CausalStore) Periodically() {
	for rid, buffer := range s.buffers {
		if !buffer.empty() && s.readyToApply(buffer.last()) {
			var u = buffer.dequeue()

			if row, ok := s.store[u.key]; !ok || row.Ts.Less(u.value.Ts) {
				s.store[u.key] = row
			}

			buffer.lastProcessed = u.value.Ts.Number

			if u.value.Ts.Number > buffer.lastProcessed { //keep up with time
				buffer.lastProcessed = u.value.Ts.Number
			}

			s.buffers[rid] = buffer
		}
	}
}
