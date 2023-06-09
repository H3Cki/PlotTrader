package trade

import (
	"fmt"
	"sync"
	"time"

	"github.com/H3Cki/TrendTrader/geometry"
	"github.com/H3Cki/TrendTrader/interval"
	"github.com/H3Cki/TrendTrader/logger"
	"github.com/google/uuid"
)

type Exchange interface {
	Name() string
	OrderPending(id any) (bool, error)
	CancelOrder(id any) error
	CreateOrder(OrderRequest) (orderID any, err error)
}

type OrderRequest struct {
	Symbol   string
	Side     string
	Price    float64
	BaseQty  float64
	QuoteQty float64
}

// BaseQuantity is a utility function that will calculate the base quantity in case only quote quantity was provided
func (o OrderRequest) BaseQuantity() float64 {
	if o.BaseQty != 0 {
		return o.BaseQty
	}

	return o.QuoteQty / o.Price
}

type FollowRequest struct {
	Symbol        string
	Interval      string
	Side          string
	BaseQuantity  float64
	QuoteQuantity float64
	Plot          geometry.Plot
}

// Follow is a structure that keeps together the order ID returned by the exchange, FollowRequest and a stop function to cancel order following
type Follow struct {
	ID       string
	OrderID  any
	Exchange string
	Request  FollowRequest
	// stop kills the internal notification channel which causes the follower to exit
	stop func()
}

type PlotFollower struct {
	exchange Exchange
	ticker   *interval.Ticker
	follows  []*Follow
	mu       sync.Mutex
}

func NewPlotFollower(exchange Exchange) *PlotFollower {
	return &PlotFollower{
		exchange: exchange,
		ticker:   interval.NewTicker(),
		follows:  []*Follow{},
		mu:       sync.Mutex{},
	}
}

// Follow starts following provided geometry, creates an order and updates it's price at every interval
func (pf *PlotFollower) Follow(req FollowRequest) (Follow, error) {
	itvDuration, err := interval.ParseDuration(req.Interval)
	if err != nil {
		return Follow{}, fmt.Errorf("unable to parse interval duration: %w", err)
	}

	orderReq := &OrderRequest{
		Symbol:   req.Symbol,
		Side:     req.Side,
		Price:    req.Plot.At(time.Now()),
		BaseQty:  req.BaseQuantity,
		QuoteQty: req.QuoteQuantity,
	}

	orderId, err := pf.exchange.CreateOrder(*orderReq)
	if err != nil {
		return Follow{}, fmt.Errorf("unable to create initial order: %w", err)
	}

	tickerC, stopTicker, err := pf.ticker.Add(interval.NextStart(time.Now(), itvDuration), itvDuration)
	if err != nil {
		return Follow{}, err
	}

	follow := &Follow{
		ID:       uuid.NewString(),
		OrderID:  orderId,
		Exchange: pf.exchange.Name(),
		Request:  req,
		stop:     stopTicker,
	}

	pf.addFollow(follow)

	go func() {
		defer func() {
			follow.stop()
		}()

		for tick := range tickerC {
			if err := pf.tickFollow(tick.Timestamp, follow, orderReq); err != nil {
				break
			}
		}
	}()

	return *follow, nil
}

// tickFollow cancels current order for a given follow and
func (pf *PlotFollower) tickFollow(t time.Time, follow *Follow, orderReq *OrderRequest) error {
	// check if order is pending, if it's not then cancel following
	pending, err := pf.exchange.OrderPending(follow.OrderID)
	if err != nil {
		logger.Errorf("unable to check if order %v is pending: %w", follow.OrderID, err)
	}

	if !pending {
		return fmt.Errorf("order %v not pending", follow.OrderID)
	}

	// cancel existing order to update the price
	if err := pf.exchange.CancelOrder(follow.OrderID); err != nil {
		return fmt.Errorf("unable to cancel order %v: %w", follow.OrderID, err)
	}

	orderReq.Price = follow.Request.Plot.At(t)

	// create new order with new price
	orderId, err := pf.exchange.CreateOrder(*orderReq)
	if err != nil {
		return fmt.Errorf("unable to create order: %w", err)
	}

	// new order was created, need to update ID
	follow.OrderID = orderId

	return nil
}

// StopFollow stops order following and removes it from the list
func (pf *PlotFollower) StopFollow(followID string) {
	for _, follow := range pf.follows {
		if follow.ID == followID {
			follow.stop()
			pf.removeFollow(followID)
		}
	}
}

func (pf *PlotFollower) Follows() []Follow {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	follows := []Follow{}

	for _, follow := range pf.follows {
		follows = append(follows, *follow)
	}

	return follows
}

func (pf *PlotFollower) addFollow(f *Follow) {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	pf.follows = append(pf.follows, f)
}

// removeFollow removes the follow from the list
func (pf *PlotFollower) removeFollow(id string) {
	pf.mu.Lock()
	defer pf.mu.Unlock()
	for i, follow := range pf.follows {
		if id == follow.ID {
			pf.follows = append(pf.follows[:i], pf.follows[i+1:]...)
			return
		}
	}
}
