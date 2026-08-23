package capacity

import (
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
	"math"
)

type Forecast struct {
	WindowMinutes      int
	ExpectedPassengers int
	Gates              int
	ServiceSeconds     float64
}

func (f Forecast) Validate() error {
	if f.WindowMinutes <= 0 || f.ExpectedPassengers < 0 || f.Gates <= 0 || f.ServiceSeconds <= 0 {
		return fmt.Errorf("%w: invalid capacity forecast", domain.ErrInvalid)
	}
	return nil
}
func (f Forecast) RequiredRate() float64 {
	return float64(f.ExpectedPassengers) / float64(f.WindowMinutes)
}
func (f Forecast) Utilization() float64 {
	return f.RequiredRate() / (60 / float64(f.ServiceSeconds) * float64(f.Gates))
}
func (f Forecast) PeakExceeded() bool { return math.Round(f.Utilization()*1000)/1000 > 1 }
