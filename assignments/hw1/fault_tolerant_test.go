package register

import (
	"github.com/dati-mipt/distributed-storage-algorithms/network"
	"math/rand"
	"sync"
	"testing"
)

const (
	CLUSTER_SIZE int = 30
	ITER_NUM     int = CLUSTER_SIZE * 5
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

	if CLUSTER_SIZE < 2 {
		t.Errorf("CLUSTER_SIZE should be at least 2, now %d", CLUSTER_SIZE)
		return
	}

	var regs []*FaultTolerantRegister
	for i := int64(0); i < int64(CLUSTER_SIZE); i++ {
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
		for i := 0; i < CLUSTER_SIZE*5; i++ {
			num := (rand.Int() / 2) % CLUSTER_SIZE
			regs[num].Write(int64(i))
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for i := 0; i < CLUSTER_SIZE*5; i++ {
			num := rand.Int() % CLUSTER_SIZE
			val := regs[num].Read()
			if val >= int64(CLUSTER_SIZE*5) || val < 0 {
				t.Errorf("wrong read value regs[%d], got: %d, max expected %d", num, val, int64(CLUSTER_SIZE*5-1))
				return
			}
		}
		wg.Done()
	}()

	wg.Wait()
	n.Wait()

	if regs[0].Read() != int64(CLUSTER_SIZE*5-1) {
		t.Errorf("wrong read value regs[%d], got: %d, expected %d", 0, regs[0].Read(), int64(CLUSTER_SIZE*5-1))
		return
	}
}
