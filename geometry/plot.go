package geometry

import (
	"time"
)

type Plot interface {
	At(time.Time) float64
}

type Point struct {
	Date  time.Time
	Price float64
}

func timeToFloat64(t time.Time) float64 {
	return float64(t.Unix())
}
