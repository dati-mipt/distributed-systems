package ftregister

import (
	"github.com/dati-mipt/distributed-storage-algorithms/network"
	"github.com/dati-mipt/distributed-storage-algorithms/util"
)

type FaultTolerantRegister struct {
	rid int64
	current util.TimestampedValue
	replicas map[int64]network.Link
}

func NewFaultTolerantRegister(rid int64) *FaultTolerantRegister{
	return &FaultTolerantRegister{
		rid:      rid,
		replicas: map[int64]network.Link{},
	}
}