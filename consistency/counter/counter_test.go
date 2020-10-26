package counter

import (
	"fmt"
	"testing"
)

type mockCounter struct {
	current int64
}

func (c *mockCounter) Inc() bool {
	c.current++
	return true
}

func (c *mockCounter) Read() int64 {
	return c.current
}

func TestGenericOperations(t *testing.T) {
	var mock = mockCounter{}
	var broadcast = BroadcastCounter{}
	var epidemic = EpidemicCounter{
		counts: replicatedCounts{},
	}

	var singleCopyIncCheck = func(counter Counter) error {
		var expected int64 = 5
		for i := int64(0); i < expected; i++ {
			counter.Inc()
		}

		var read = counter.Read()
		if read != expected {
			return fmt.Errorf("wrong counter value, got: %v, expected %v", read, expected)
		}

		return nil
	}

	for _, c := range []Counter{&mock, &broadcast, &epidemic} {
		if err := singleCopyIncCheck(c); err != nil {
			t.Errorf("failed single copy API test for %T: %v", c, err)
		}
	}
}
