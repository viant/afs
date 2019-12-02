package base

import (
	"math/rand"
	"time"
)

//Retry represents abstraction holding sleep duration between retries (back-off)
type Retry struct {
	Count      int
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
	duration   time.Duration
}

// Pause returns the next time.Duration that the caller should use to backoff.
func (b *Retry) Pause() time.Duration {
	if b.Initial == 0 {
		b.Initial = time.Second
	}
	if b.duration == 0 {
		b.duration = b.Initial
	}
	if b.Max == 0 {
		b.Max = 30 * time.Second
	}
	if b.Multiplier < 1 {
		b.Multiplier = 2
	}

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := time.Duration(1 + rnd.Int63n(int64(b.duration)))
	b.duration = time.Duration(float64(b.duration) * b.Multiplier)
	if b.duration > b.Max {
		b.duration = b.Max
	}
	return result
}

//NewRetry creates a retry
func NewRetry() *Retry {
	return &Retry{}
}
