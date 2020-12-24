package ftr

import (
	"github.com/dati-mipt/distributed-storage-algorithms/network"
	"github.com/dati-mipt/distributed-storage-algorithms/util"
	"fmt"
)

type FTRegister struct {
	rid      int64
	current  util.TimestampedValue

	replicas map[int64]network.Link
}

func NewFTRegister(rid int64) *FTRegister {
	return &FTRegister{
		rid:      rid,
		replicas: map[int64]network.Link{},
	}
}

func (r *FTRegister) SingleWrite(value util.TimestampedValue) bool {
	for _, link := range r.replicas {
		link.AsyncMessage(value);
	}
	// In real ftr we must wait len(r.replicas)/2 + 1 recvs.
	// It's nessery to anderstand did our system crash,
	// but we will check this in SingleRead() only!
	// Because of this in Write() and Read() we will check this.
	return true
}

func (r *FTRegister) SingleRead() int64 {
	system_size := len(r.replicas)
	// TODO waiteing for system_size / 2 + 1 recvs in read.
}


func (r *FTRegister) Write(value int64) bool {
	// TODO
	// write is SingleRead + SingleWrite is this order
	return true
}

func (r *FTRegister) Read() int64 {
	// TODO 
	// write is SingleRead + SingleWrite is this order
	return 0
}

func (r *FTRegister) Introduce(rid int64, link network.Link) {
	if link != nil {
		r.replicas[rid] = link
	}
}

func (r *FTRegister) Receive(rid int64, msg interface{}) interface{} {
	if update, ok := msg.(util.TimestampedValue); ok {
		if r.current.Ts.Less(update.Ts) {
			r.current = update
		}
	}
	return nil
}
