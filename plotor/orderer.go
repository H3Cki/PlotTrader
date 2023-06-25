package plotor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/H3Cki/Plotor/geometry"
	"github.com/H3Cki/Plotor/logger"
)

type Client interface {
	CreateOrder(ctx context.Context, orderData any, price float64) (order ClientOrder, err error)
	GetOrder(ctx context.Context, order ClientOrder) (newOrder ClientOrder, err error)
	UpdateOrderPrice(ctx context.Context, order ClientOrder, price float64) (newOrder ClientOrder, err error)
	CancelOrder(ctx context.Context, order ClientOrder) (err error)
}

// PlotOrderer is responsible for managing plot orders using given client
type PlotOrderer struct {
	client     Client
	plotOrders map[string]*PlotOrder
	mu         *sync.Mutex
}

func NewPlotOrderer(c Client) *PlotOrderer {
	return &PlotOrderer{
		client:     c,
		plotOrders: map[string]*PlotOrder{},
		mu:         &sync.Mutex{},
	}
}

func (p *PlotOrderer) Client() Client {
	return p.client
}

// Get returns a copy of the PlotOrder with up-to-date Order fetched from the client
func (p *PlotOrderer) Get(ctx context.Context, plotOrderID string) (*PlotOrder, error) {
	p.mu.Lock()
	po, ok := p.plotOrders[plotOrderID]
	p.mu.Unlock()

	if !ok {
		return nil, fmt.Errorf("plot order %s not found", plotOrderID)
	}

	order, err := p.client.GetOrder(ctx, po.Order)
	if err != nil {
		return nil, fmt.Errorf("error getting order from client: %w", err)
	}

	return &PlotOrder{
		ID:       po.ID,
		IsActive: po.IsActive,
		Plot:     po.Plot,
		Interval: po.Interval,
		Order:    order,
		LastTick: po.LastTick,
		stopC:    make(chan struct{}),
	}, nil
}

// Create creates a plot order and updates it continuously until the plot goes out of range or there is an error
func (p *PlotOrderer) Create(ctx context.Context, orderData any, plot geometry.Plot, interval time.Duration) (*PlotOrder, error) {
	price, err := plot.At(time.Now())
	if err != nil {
		return nil, err
	}

	order, err := p.client.CreateOrder(ctx, orderData, price)
	if err != nil {
		return nil, err
	}

	po := NewPlotOrder(order, plot, interval)

	p.mu.Lock()
	p.plotOrders[po.ID] = po
	p.mu.Unlock()

	go func() {
		if err := po.RunNextInterval(p.handler(ctx, po)); err != nil {
			logger.Errorf("error updating plot order: %v", err)
		}
	}()

	return po, nil
}

// Stop stops updating the plot order, if cancelOrder is true then it also cancels the order on the exchange
func (p *PlotOrderer) Stop(ctx context.Context, plotOrderID string, cancelOrder bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cancelOrder(ctx, plotOrderID, cancelOrder)
}

func (p *PlotOrderer) StopAll(ctx context.Context, cancelOrder bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for id := range p.plotOrders {
		if err := p.cancelOrder(ctx, id, cancelOrder); err != nil {
			return err
		}
	}

	return nil
}

func (p *PlotOrderer) handler(ctx context.Context, po *PlotOrder) Handler {
	return func(order ClientOrder, price float64) error {
		newOrder, err := p.client.UpdateOrderPrice(ctx, order, price)
		if err != nil {
			return err
		}

		po.Order = newOrder

		return err
	}
}

func (p *PlotOrderer) cancelOrder(ctx context.Context, plotOrderID string, cancelOrder bool) error {
	po, ok := p.plotOrders[plotOrderID]
	if !ok {
		return fmt.Errorf("order %s does not exit", plotOrderID)
	}

	po.Stop()
	//delete(p.plotOrders, po.ID)
	if cancelOrder {
		return p.client.CancelOrder(ctx, po.Order)
	}

	return nil
}
