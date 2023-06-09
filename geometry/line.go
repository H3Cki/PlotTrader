package geometry

import (
	"math"
	"time"
)

type Line struct {
	A, B float64
}

func (l *Line) At(date time.Time) float64 {
	y := l.A*timeToFloat64(date) + l.B

	return y
}

func NewLine(p1, p2 Point) *Line {
	p1DateFloat := timeToFloat64(p1.Date)
	p2DateFloat := timeToFloat64(p2.Date)

	a := (p2.Price - p1.Price) / (p2DateFloat - p1DateFloat)
	b := p1.Price - (a * p1DateFloat)

	l := &Line{
		A: a,
		B: b,
	}

	return l
}

// Straight line on semi-logarighmic (x, log10) graph
type LogLine struct {
	M, K, Xoffset float64
}

func NewLogLine(p0, p1 Point) *LogLine {
	xOffset := timeToFloat64(p0.Date)

	x0 := 0.0
	x1 := timeToFloat64(p1.Date) - xOffset

	y0 := p0.Price
	y1 := p1.Price

	m := (math.Log10(y1) - math.Log10(y0)) / (x1 - x0)

	return &LogLine{
		M:       m,
		K:       y0,
		Xoffset: xOffset,
	}
}

func (l *LogLine) At(date time.Time) float64 {
	x := timeToFloat64(date) - l.Xoffset

	return l.K * math.Pow(10, l.M*x)
}

type HorizontalLine struct {
	Price float64
}

func (l *HorizontalLine) At(date time.Time) float64 {
	return l.Price
}

func NewHorizontalLine(price float64) *HorizontalLine {
	return &HorizontalLine{price}
}
