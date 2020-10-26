package store

import (
	"fmt"
	"testing"

	"github.com/dati-mipt/distributed-algorithms/util"
)

type mockStore struct {
	data map[int64]int64
}

func (s *mockStore) Write(key int64, value int64) bool {
	if s.data == nil {
		s.data = map[int64]int64{}
	}
	s.data[key] = value
	return true
}

func (s *mockStore) Read(key int64) int64 {
	if s.data == nil {
		return 0
	}
	return s.data[key]
}

func TestGenericOperations(t *testing.T) {
	var mock = mockStore{data: map[int64]int64{}}
	var evential = EventualStore{store: map[int64]util.TimestampedValue{}}
	var causal = CausalStore{
		store:   map[int64]util.TimestampedValue{},
		buffers: map[int64]inBuffer{},
		deps:    map[int64]util.Timestamp{},
	}

	var singleCopyStoreCheck = func(store Store) error {
		var key int64 = 4
		var value int64 = 6
		store.Write(key, value)

		var read = store.Read(key)
		if read != value {
			return fmt.Errorf("wrong store value, got: %v, expected %v", read, value)
		}

		return nil
	}

	for _, c := range []Store{&mock, &evential, &causal} {
		if err := singleCopyStoreCheck(c); err != nil {
			t.Errorf("failed single copy API test for %T: %v", c, err)
		}
	}
}
