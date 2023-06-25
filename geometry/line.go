package geometry

import (
	"errors"
	"math"
	"time"
)

type Point struct {
	Date  time.Time
	Price float64
}

type Line struct {
	A, B                  float64
	LeftLimit, RightLimit time.Time
}

func NewLine(p0, p1 Point, extendLeft, extendRight bool) (*Line, error) {
	if p0.Date.Equal(p1.Date) {
		return nil, errors.New("error creating new line: both points have the same date")
	}

	sorted := sortPoints(p0, p1)
	p0 = sorted[0]
	p1 = sorted[1]

	p0DateFloat := timeToFloat64(p0.Date)
	p1DateFloat := timeToFloat64(p1.Date)

	a := (p1.Price - p0.Price) / (p1DateFloat - p0DateFloat)
	b := p0.Price - (a * p0DateFloat)

	l := &Line{
		A: a,
		B: b,
	}

	if !extendLeft {
		l.LeftLimit = p0.Date
	}

	if !extendRight {
		l.RightLimit = p1.Date
	}

	return l, nil
}

func (l *Line) At(date time.Time) (float64, error) {
	if !lineInRange(date, l.LeftLimit, l.RightLimit) {
		return 0, ErrOutOfRange
	}

	return l.A*timeToFloat64(date) + l.B, nil
}

// Straight line on semi-logarighmic (x, log10) graph
type LogLine struct {
	M, K, Xoffset         float64
	LeftLimit, RightLimit time.Time
}

func NewLogLine(p0, p1 Point, extendLeft, extendRight bool) (*LogLine, error) {
	if p0.Date.Equal(p1.Date) {
		return nil, errors.New("error creating new line: both points have the same date")
	}

	sorted := sortPoints(p0, p1)
	p0 = sorted[0]
	p1 = sorted[1]

	xOffset := timeToFloat64(p0.Date)

	x0 := 0.0
	x1 := timeToFloat64(p1.Date) - xOffset

	y0 := p0.Price
	y1 := p1.Price
	m := (math.Log10(y1) - math.Log10(y0)) / (x1 - x0)

	l := &LogLine{
		M:       m,
		K:       y0,
		Xoffset: xOffset,
	}

	if !extendLeft {
		l.LeftLimit = p0.Date
	}

	if !extendRight {
		l.RightLimit = p1.Date
	}

	return l, nil
}

func (l *LogLine) At(date time.Time) (float64, error) {
	if !lineInRange(date, l.LeftLimit, l.RightLimit) {
		return 0, ErrOutOfRange
	}

	x := timeToFloat64(date) - l.Xoffset
	return l.K * math.Pow(10, l.M*x), nil
}

// lineInRange performs a [leftLimit, rightLimit) check
func lineInRange(t, leftLimit, rightLimit time.Time) bool {
	return (!t.Before(leftLimit) || leftLimit.IsZero()) && (t.Before(rightLimit) || rightLimit.IsZero())
}
