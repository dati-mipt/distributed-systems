package store

type Store interface {
	Write(key int64, value int64) bool
	Read(key int64) int64
}
