package hw1

import (
	"github.com/dati-mipt/distributed-systems/network"
)

type FaultTolerantRegister struct {
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister {
	return &FaultTolerantRegister{}
}

func (r *FaultTolerantRegister) Write(value int64) bool {
	return false
}

func (r *FaultTolerantRegister) Read() int64 {
	return 0
}

func (r *FaultTolerantRegister) Introduce(rid int64, link network.Link) {
}

func (r *FaultTolerantRegister) Receive(rid int64, msg interface{}) interface{} {
	return nil
}
