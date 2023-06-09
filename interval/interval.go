package interval

import "time"

var predefinedIntervals = map[string]time.Duration{
	"1d": 24 * time.Hour,
	"2d": 2 * 24 * time.Hour,
	"3d": 3 * 24 * time.Hour,
	"4d": 4 * 24 * time.Hour,
	"5d": 5 * 24 * time.Hour,
	"6d": 6 * 24 * time.Hour,
	"1w": 7 * 24 * time.Hour,
	"2w": 14 * 24 * time.Hour,
	"1M": 30 * 24 * time.Hour,
}

// ParseDuration adds more units on top of time.ParseDuration():
// 1d, 2d, 3d, 4d, 5d, 6d, 1w, 2w, 1M.
// These units are added for convenience and cannot be combined e.g. ParseDuration("1d12h") or ParseDuration("1w1d") wont work.
func ParseDuration(itv string) (time.Duration, error) {
	if d, ok := predefinedIntervals[itv]; ok {
		return d, nil
	}

	d, err := time.ParseDuration(itv)
	if err != nil {
		return 0, err
	}

	return d, nil
}

func NextStart(now time.Time, every time.Duration) time.Time {
	now = now.In(time.UTC)

	itvSeconds := int64(every.Seconds())
	div := now.Unix() / itvSeconds

	prevStartSeconds := div * itvSeconds
	nextStartSeconds := prevStartSeconds + itvSeconds

	return time.Unix(nextStartSeconds, 0).In(time.UTC)

}

func Start(now time.Time, every time.Duration) time.Time {
	now = now.In(time.UTC)

	itvSeconds := int64(every.Seconds())
	div := now.Unix() / itvSeconds

	prevStartSeconds := div * itvSeconds

	return time.Unix(prevStartSeconds, 0).In(time.UTC)
}
