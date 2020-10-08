package counter

type Counter interface {
	Inc() bool
	Read() int64
}
