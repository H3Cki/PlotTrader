package geometry_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/H3Cki/PlotTrader/geometry"
	"github.com/stretchr/testify/assert"
)

func TestFollowerStart(t *testing.T) {
	tests := []struct {
		plot           geometry.Plot
		interval       time.Duration
		expectedPrices []float64
	}{
		{
			interval:       time.Second,
			plot:           geometry.NewLine(geometry.Point{Date: time.Now(), Price: 0}, geometry.Point{Date: time.Now().Add(3 * time.Second), Price: 3}),
			expectedPrices: []float64{0, 1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.expectedPrices), func(t *testing.T) {
			follower := geometry.NewPlotFollower(tt.plot, tt.interval)
			err := follower.Start()
			assert.NoError(t, err)

			defer func() {
				follower.Stop()
			}()

			results := []float64{}

			deadline := time.After(time.Duration(len(tt.expectedPrices)) * tt.interval)

		loop:
			for {
				select {
				case tick := <-follower.TickerC():
					results = append(results, tick.Price)
					if len(results) == len(tt.expectedPrices) {
						break
					}
				case <-deadline:
					break loop
				}
			}

			assert.EqualValues(t, tt.expectedPrices, results)
		})
	}
}
