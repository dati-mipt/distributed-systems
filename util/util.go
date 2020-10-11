package util

type Timestamp struct {
	Number int64
	Rid    int64
}

func (t Timestamp) Less(o Timestamp) bool {
	if t.Number < o.Number || (t.Number == o.Number && t.Rid < o.Rid) {
		return true
	}

	return false
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
