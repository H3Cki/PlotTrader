package geometry

import "time"

type Offsetter interface {
	Offset(v float64) float64
}

// AbsoluteOffset offsets the value by another value
type AbsoluteOffset struct {
	Value float64
}

func NewAbsoluteOffset(value float64) *AbsoluteOffset {
	return &AbsoluteOffset{value}
}

func (p *AbsoluteOffset) Offset(v float64) float64 {
	return v + p.Value
}

// PercentageOffset offsets the value by a percentage of it
type PercentageOffset struct {
	Percentage float64
}

func NewPercentageOffset(value float64) *PercentageOffset {
	return &PercentageOffset{value}
}

func (p *PercentageOffset) Offset(v float64) float64 {
	return v + (v * p.Percentage)
}

// OffsetPlot wraps a Plot and applies an offset to Plot's returned value
type OffsetPlot struct {
	Offsetter Offsetter
	Plot      Plot
}

func NewOffsetPlot(plot Plot, offset Offsetter) *OffsetPlot {
	return &OffsetPlot{
		Offsetter: offset,
		Plot:      plot,
	}
}

func (o *OffsetPlot) At(t time.Time) float64 {
	v := o.Plot.At(t)
	return o.Offsetter.Offset(v)
}
