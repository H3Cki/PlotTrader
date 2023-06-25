package geometry

import (
	"errors"
	"sort"
	"time"
)

var ErrOutOfRange = errors.New("out of plot's range")

type Plot interface {
	At(time.Time) (float64, error)
	// Marshal() ([]byte, error)
}

func timeToFloat64(t time.Time) float64 {
	return float64(t.Unix())
}

func sortPoints(points ...Point) []Point {
	sort.Slice(points, func(i, j int) bool { return points[i].Date.Before(points[j].Date) })
	return points
}
