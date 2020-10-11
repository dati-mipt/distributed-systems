package store

import "github.com/dati-mipt/consistency-algorithms/util"

type Store interface {
	Write(key int64, value int64) bool
	Read(key int64) int64
}

type timedRow struct {
	ts  util.Timestamp
	val int64
}
