package hw1

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/dati-mipt/distributed-systems/network"
)

func TestFaultTolerantRegister(t *testing.T) {
	var n = network.NewReliableNetwork()

	var regs []*FaultTolerantRegister
	for i := int64(0); i < 2; i++ {
		var reg = NewFaultTolerantRegister(i)
		n.Register(i, reg)
		regs = append(regs, reg)
	}

	if regs == nil {
		t.Error("syntax check failed")
		return
	}

	go n.Route()

	if regs[0].Read() != 0 {
		t.Errorf("wrong read value regs[%d], got: %d, expected %d", 0, regs[0].Read(), 0)
		return
	}

	n.Wait()
	regs[0].Write(4)

	n.Wait()
	if regs[0].Read() != 4 {
		t.Errorf("wrong read value regs[%d], got: %d, expected %d", 0, regs[0].Read(), 4)
		return
	}

	n.Wait()
	if regs[1].Read() != 4 {
		t.Errorf("wrong read value regs[%d], got: %d, expected %d", 1, regs[1].Read(), 4)
		return
	}
}

func TestFaultTolerantRegisterDieHard(t *testing.T) {
	var n = network.NewReliableNetwork()

	var clusterSize = 30
	var iterationsNumber = clusterSize * 5

	var regs []*FaultTolerantRegister
	for i := int64(0); i < int64(clusterSize); i++ {
		var reg = NewFaultTolerantRegister(i)
		n.Register(i, reg)
		regs = append(regs, reg)
	}

	if regs == nil {
		t.Error("syntax check failed")
		return
	}

	go n.Route()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for i := 0; i < iterationsNumber; i++ {
			num := (rand.Int() / 2) % clusterSize
			regs[num].Write(int64(i))
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for i := 0; i < iterationsNumber; i++ {
			num := rand.Int() % clusterSize
			val := regs[num].Read()
			if val >= int64(iterationsNumber) || val < 0 {
				t.Errorf("wrong read value regs[%d], got: %d, max expected %d", num, val, int64(iterationsNumber-1))
				return
			}
		}
		wg.Done()
	}()

	wg.Wait()
	n.Wait()

	if regs[0].Read() != int64(iterationsNumber-1) {
		t.Errorf("wrong read value regs[%d], got: %d, expected %d", 0, regs[0].Read(), int64(iterationsNumber-1))
		return
	}
}

func TestFaultTolerantRegisterRealtime(t *testing.T) {
	var n = network.NewReliableNetwork()
	var clusterSize = 51
	var nIters = 50

	var regs []*FaultTolerantRegister
	for i := int64(0); i < int64(clusterSize); i++ {
		var reg = NewFaultTolerantRegister(i)
		n.Register(i, reg)
		regs = append(regs, reg)
	}

	go n.Route()

	for i := 0; i < nIters; i++ {
		var val int64
		for j := 0; j < 5; j++ {
			val = int64(rand.Int())
			regs[rand.Int() % clusterSize].Write(val)
		}
		if regs[rand.Int() % clusterSize].Read() != val {
			t.Errorf("wrong read value")
			return
		}
	}
}

func TestFaultTolerantRegisterIncOrder(t *testing.T) {
	var n = network.NewReliableNetwork()
	var clusterSize = 51
	var nIters = 50

	var regs []*FaultTolerantRegister
	for i := int64(0); i < int64(clusterSize); i++ {
		var reg = NewFaultTolerantRegister(i)
		n.Register(i, reg)
		regs = append(regs, reg)
	}

	go n.Route()

	init := int64(rand.Int())
	regs[0].Write(init)

	for i := 0; i < nIters; i++ {
		id := rand.Int() % clusterSize
		regs[id].Write(regs[id].Read() + 1)
	}

	if regs[0].Read() != init + int64(nIters) {
		t.Errorf("test failed")
	}
}
