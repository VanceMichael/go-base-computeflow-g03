package worker

import "time"

type Backoff struct {
	Initial time.Duration
	Max     time.Duration
}

func (b Backoff) Delay(attempt int) time.Duration {
	if attempt < 1 {
		return 0
	}
	d := b.Initial
	for n := 1; n < attempt; n++ {
		d *= 2
		if d >= b.Max {
			return b.Max
		}
	}
	if d > b.Max {
		return b.Max
	}
	return d
}

type RetryPolicy struct {
	MaxAttempts int
	Backoff     Backoff
}

func (p RetryPolicy) Retry(attempt int) bool { return attempt < p.MaxAttempts }
