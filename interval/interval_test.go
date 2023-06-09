package interval_test

import (
	"testing"
	"time"

	"github.com/H3Cki/TrendTrader/interval"
	"github.com/stretchr/testify/assert"
)

func TestNextStart(t *testing.T) {
	tests := []struct {
		now  time.Time
		itv  time.Duration
		next time.Time
	}{
		{time.Date(2023, 1, 15, 12, 30, 30, 0, time.UTC), time.Second, time.Date(2023, 1, 15, 12, 30, 31, 0, time.UTC)},
		{time.Date(2023, 1, 15, 12, 30, 30, 0, time.UTC), time.Minute, time.Date(2023, 1, 15, 12, 31, 0, 0, time.UTC)},
		{time.Date(2023, 1, 15, 12, 30, 30, 0, time.UTC), 15 * time.Minute, time.Date(2023, 1, 15, 12, 45, 0, 0, time.UTC)},
		{time.Date(2023, 1, 15, 12, 30, 30, 0, time.UTC), time.Hour, time.Date(2023, 1, 15, 13, 0, 0, 0, time.UTC)},
		{time.Date(2023, 1, 15, 12, 30, 30, 0, time.UTC), 4 * time.Hour, time.Date(2023, 1, 15, 16, 0, 0, 0, time.UTC)},
		{time.Date(2023, 1, 15, 23, 30, 30, 0, time.UTC), 4 * time.Hour, time.Date(2023, 1, 16, 0, 0, 0, 0, time.UTC)},
		{time.Date(2023, 1, 15, 23, 30, 30, 0, time.UTC), 24 * time.Hour, time.Date(2023, 1, 16, 0, 0, 0, 0, time.UTC)},
	}

	for _, test := range tests {
		t.Run(test.now.String()+test.itv.String(), func(t *testing.T) {
			next := interval.NextStart(test.now, test.itv)
			assert.Equal(t, test.next, next)
		})
	}
}

func TestStart(t *testing.T) {
	t.Fail()

	tests := []struct {
		now  time.Time
		itv  time.Duration
		next time.Time
	}{
		{time.Date(2023, 1, 15, 12, 30, 30, 0, time.UTC), time.Second, time.Date(2023, 1, 15, 12, 30, 31, 0, time.UTC)},
		{time.Date(2023, 1, 15, 12, 30, 30, 0, time.UTC), time.Minute, time.Date(2023, 1, 15, 12, 31, 0, 0, time.UTC)},
		{time.Date(2023, 1, 15, 12, 30, 30, 0, time.UTC), 15 * time.Minute, time.Date(2023, 1, 15, 12, 45, 0, 0, time.UTC)},
		{time.Date(2023, 1, 15, 12, 30, 30, 0, time.UTC), time.Hour, time.Date(2023, 1, 15, 13, 0, 0, 0, time.UTC)},
		{time.Date(2023, 1, 15, 12, 30, 30, 0, time.UTC), 4 * time.Hour, time.Date(2023, 1, 15, 16, 0, 0, 0, time.UTC)},
		{time.Date(2023, 1, 15, 23, 30, 30, 0, time.UTC), 4 * time.Hour, time.Date(2023, 1, 16, 0, 0, 0, 0, time.UTC)},
		{time.Date(2023, 1, 15, 23, 30, 30, 0, time.UTC), 24 * time.Hour, time.Date(2023, 1, 16, 0, 0, 0, 0, time.UTC)},
	}

	for _, test := range tests {
		t.Run(test.now.String()+test.itv.String(), func(t *testing.T) {
			next := interval.NextStart(test.now, test.itv)
			assert.Equal(t, test.next, next)
		})
	}
}
