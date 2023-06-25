package plotor

import (
	"time"

	"github.com/H3Cki/Plotor/geometry"
	"github.com/google/uuid"
)

type ClientOrder interface {
	Details() (map[string]any, error)
}

type Handler func(order ClientOrder, price float64) error

// PlotOrder aggregates neccessary information and uses it to update the order
type PlotOrder struct {
	ID       string
	IsActive bool
	Plot     geometry.Plot
	Interval time.Duration
	Order    ClientOrder
	LastTick time.Time
	stopC    chan struct{}
	ticker   *time.Ticker
}

func NewPlotOrder(order ClientOrder, plot geometry.Plot, interval time.Duration) *PlotOrder {
	return &PlotOrder{
		ID:       uuid.NewString(),
		Plot:     plot,
		Interval: interval,

		Order: order,
		stopC: make(chan struct{}),
	}
}

// Run starts ticking the price on the plot every given interval and passing it to the handler
func (po *PlotOrder) Run(handler Handler) error {
	return po.run(time.Now(), handler)
}

// RunNextInterval waits until the start of the next interval to start running
func (po *PlotOrder) RunNextInterval(handler Handler) error {
	t := <-time.After(time.Until(NextIntervalStart(time.Now(), po.Interval)))
	return po.run(t, handler)
}

func (po *PlotOrder) run(t time.Time, handler Handler) error {
	po.ticker = time.NewTicker(po.Interval)
	po.IsActive = true
	defer func() {
		po.ticker.Stop()
		po.IsActive = false
	}()

	for {
		price, err := po.Plot.At(t)
		if err != nil {
			return err
		}

		if err := handler(po.Order, price); err != nil {
			return err
		}

		po.LastTick = t

		select {
		case <-po.stopC:
			return nil
		case tick := <-po.ticker.C:
			t = tick
		}
	}
}

func (po *PlotOrder) Stop() {
	close(po.stopC)
}
