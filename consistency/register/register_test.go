package register

import (
	"fmt"
	"testing"
)

type mockRegister struct {
	current int64
}

func (r *mockRegister) Write(value int64) bool {
	r.current = value
	return true
}

func (r *mockRegister) Read() int64 {
	return r.current
}

func TestGenericOperations(t *testing.T) {
	var mock = mockRegister{}
	var epidemic = EpidemicRegister{}

	var singleCopyRegisterCheck = func(reg Register) error {
		var expected int64 = 5
		reg.Write(expected)

		var read = reg.Read()
		if read != expected {
			return fmt.Errorf("wrong register value, got: %v, expected %v", read, expected)
		}

		return nil
	}

	for _, c := range []Register{&mock, &epidemic} {
		if err := singleCopyRegisterCheck(c); err != nil {
			t.Errorf("failed single copy API test for %T: %v", c, err)
		}
	}
}
