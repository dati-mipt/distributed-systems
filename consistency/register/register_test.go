package register

import (
	"fmt"
	"github.com/dati-mipt/distributed-storage-algorithms/network"
	"testing"
	"time"
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
	var epidemic = NewEpidemicRegister(0)

	var singleCopyRegisterCheck = func(reg Register) error {
		var expected int64 = 5
		reg.Write(expected)

		var read = reg.Read()
		if read != expected {
			return fmt.Errorf("wrong register value, got: %v, expected %v", read, expected)
		}

		return nil
	}

	for _, c := range []Register{&mock, epidemic} {
		if err := singleCopyRegisterCheck(c); err != nil {
			t.Errorf("failed single copy API test for %T: %v", c, err)
		}
	}
}

func TestEpidemicRegisterReplicaSet(t *testing.T) {
	var n = network.NewReliableNetwork()

	var regs []*EpidemicRegister
	for i := int64(0); i < 2; i++ {
		var reg = NewEpidemicRegister(i)
		n.Register(i, reg)
		regs = append(regs, reg)
	}

	if regs == nil {
		t.Error("syntax check failed")
		return
	}

	go n.Route()

	if regs[0].Read() != 0 {
		t.Errorf("wrong read value, got: %d, expected %d", regs[0].Read(), 0)
		return
	}

	regs[0].Write(4)
	regs[0].Periodically()

	time.Sleep(time.Millisecond)

	if regs[1].Read() != 4 {
		t.Errorf("wrong read value, got: %d, expected %d", regs[1].Read(), 4)
		return
	}
}
