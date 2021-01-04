package util

type Timestamp struct {
	Number int64
	Rid    int64
}

func (t Timestamp) Less(o Timestamp) bool {
	return t.Number < o.Number || (t.Number == o.Number && t.Rid < o.Rid)
}

type TimestampedValue struct {
	Val int64
	Ts  Timestamp
}

func (t *TimestampedValue) Store(o TimestampedValue) {
	if o.Ts.Number >= t.Ts.Number {
		*t = o
	}
}

// Max returns the larger of x or y.
func Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

type Ticker interface {
	Periodically()
}

type Role int

const (
	Client Role = iota
	Server
)

func (r Role) String() string {
	return [...]string{"Client", "Server"}[r]
}
