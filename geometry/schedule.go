package geometry

import (
	"time"

	"github.com/H3Cki/Plotor/logger"
)

// Schedule is a Plot wrapper that allows it to be valid only in certain periods of time.
// The range is [Since, Until), time constraints are considered active only when they are non-zero,
// meaning that if Since is zero and Until is not, only Until constraint will be applied.
type Schedule struct {
	Since, Until time.Time
	Plot         Plot
}

// NewSchedule is a constructor for Schedule, sets since and until times to UTC timezone
func NewSchedule(since, until time.Time, plot Plot) *Schedule {
	logger.Infof("since: %s | until: %s", since.String(), until.String())
	return &Schedule{Since: since, Until: until, Plot: plot}
}

func (v *Schedule) At(t time.Time) (float64, error) {
	if !v.InRange(t) {
		return 0, ErrOutOfRange
	}

	return v.Plot.At(t)
}

func (v *Schedule) InRange(t time.Time) bool {
	logger.Infof("\n since: %s | until: %s\nt: %s", v.Since.String(), v.Until.String(), t.String())

	return (!t.Before(v.Since) || v.Since.IsZero()) && (t.Before(v.Until) || v.Until.IsZero())
}
