package timepolicy

import "time"

type Policy struct {
	Location *time.Location
	Grace    time.Duration
}

func New(zone *time.Location, grace time.Duration) Policy {
	return Policy{Location: zone, Grace: grace}
}
func (p Policy) StartOfDay(t time.Time) time.Time {
	local := t.In(p.Location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, p.Location)
}
func (p Policy) Window(t time.Time) (time.Time, time.Time) {
	start := p.StartOfDay(t)
	return start, start.Add(24 * time.Hour)
}
func (p Policy) Expired(deadline, now time.Time) bool { return !now.Before(deadline.Add(p.Grace)) }
func (p Policy) SameBusinessDay(a, b time.Time) bool {
	sa, _ := p.Window(a)
	sb, _ := p.Window(b)
	return sa.Equal(sb)
}
