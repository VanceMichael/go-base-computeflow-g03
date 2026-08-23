package timepolicy_test

import (
	"github.com/VanceMichael/computeflow/internal/timepolicy"
	"testing"
	"time"
)

func TestWindowFollowsBusinessTimezone(t *testing.T) {
	zone, _ := time.LoadLocation("Asia/Shanghai")
	p := timepolicy.New(zone, 0)
	from, to := p.Window(time.Date(2026, 8, 23, 23, 30, 0, 0, time.UTC))
	if from.Hour() != 0 || to.Sub(from) != 24*time.Hour {
		t.Fatalf("%v %v", from, to)
	}
}
func TestExpiredHonorsGracePeriod(t *testing.T) {
	p := timepolicy.New(time.UTC, time.Minute)
	deadline := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if p.Expired(deadline, deadline.Add(30*time.Second)) {
		t.Fatal("within grace expired")
	}
	if !p.Expired(deadline, deadline.Add(61*time.Second)) {
		t.Fatal("past grace still active")
	}
}
func TestSameBusinessDayUsesConfiguredZone(t *testing.T) {
	p := timepolicy.New(time.FixedZone("port", 8*3600), 0)
	a := time.Date(2026, 8, 23, 14, 30, 0, 0, time.UTC)
	b := time.Date(2026, 8, 23, 15, 0, 0, 0, time.UTC)
	if !p.SameBusinessDay(a, b) {
		t.Fatal("same local day should match")
	}
}
