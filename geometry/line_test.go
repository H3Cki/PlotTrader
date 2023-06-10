package geometry_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/H3Cki/PlotTrader/geometry"
	"github.com/stretchr/testify/assert"
)

func TestLine_At(t *testing.T) {
	tests := []struct {
		p1, p2 geometry.Point
		x      time.Time
		y      float64
	}{
		{
			p1: geometry.Point{time.Unix(0, 0), 0},
			p2: geometry.Point{time.Unix(1, 0), 1},
			x:  time.Unix(2, 0),
			y:  2,
		},
		{
			p1: geometry.Point{time.Unix(0, 0), 0},
			p2: geometry.Point{time.Unix(1, 0), 1},
			x:  time.Unix(-500, 0),
			y:  -500,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			line := geometry.NewLine(test.p1, test.p2)
			y := line.At(test.x)

			assert.Equal(t, test.y, y)
		})
	}
}

func TestLogLine_At(t *testing.T) {
	// Hard math
	t.Fail()
}

func TestHorizontalLine_At(t *testing.T) {
	for price := -1.0; price < 1.0; price += 0.01 {
		t.Run(fmt.Sprint(price), func(t *testing.T) {
			line := geometry.NewHorizontalLine(price)
			y := line.At(time.Now().UTC().Add(time.Duration(price) * time.Hour))

			assert.Equal(t, price, y)
		})
	}
}
