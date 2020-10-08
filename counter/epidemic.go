package counter

import "github.com/dati-mipt/consistency-algorithms/util"

type IntMapReplica interface {
	IntMapMessage(msg map[int64]int64)
}

type EpidemicCounter struct {
	rid    int64
	counts map[int64]int64

	replicas []IntMapReplica
}

func (c EpidemicCounter) Inc() bool {
	c.counts[c.rid] = c.counts[c.rid] + 1
	return true
}

func (c EpidemicCounter) Read() int64 {
	var sum = int64(0)
	for _, v := range c.counts {
		sum += v
	}
	return sum
}

func (c EpidemicCounter) IntMapMessage(msg map[int64]int64) {
	for k, v := range msg {
		c.counts[k] = util.Max(c.counts[k], v)
	}
}

func (c EpidemicCounter) Periodically() {
	for _, r := range c.replicas {
		r.IntMapMessage(c.counts)
	}
}
