package trade

import (
	"fmt"
	"time"

	"github.com/H3Cki/TrendTrader/geometry"
	"github.com/H3Cki/TrendTrader/logger"
	"github.com/google/uuid"
)

type Exchange interface {
	OrderPending(id any) (bool, error)
	GetOrder(id any) (*Order, error)
	CancelOrder(id any) error
	CreateOrder(OrderRequest) (*Order, error)
}

// Order is used to store details of an order created by an exchange.
// Price and quantities might differ from OrderRequest due to filters on exchanges.
// For example the OrderRequest has a Price of 1.2345,
// but the exchange requires the price to be at most 2 decimal places long,
// so the price must be adjusted to 1.23.
type Order struct {
	ID       any
	Symbol   string
	Side     string
	Type     string
	Price    float64
	BaseQty  float64
	QuoteQty float64

	Leverage      float64
	TimeInForce   string
	StopPrice     float64
	TrailingDelta float64
}

// OrderRequest is a structure that's passed to exchanges, so that they can create an order based on it.
// It contains rough price and quantity values unadjusted for the needs of different exchanges.
type OrderRequest struct {
	Symbol   string
	Side     string
	Type     string
	Price    float64
	BaseQty  float64
	QuoteQty float64

	Leverage      float64
	TimeInForce   string
	StopPrice     float64
	TrailingDelta float64
}

// ToOrder is a utility function that creates an order and assigns it's corresponding field values from the request
func (o *OrderRequest) ToOrder() *Order {
	return &Order{
		Symbol:        o.Symbol,
		Side:          o.Side,
		Type:          o.Type,
		Price:         o.Price,
		BaseQty:       o.BaseQty,
		QuoteQty:      o.QuoteQty,
		Leverage:      o.Leverage,
		TimeInForce:   o.TimeInForce,
		StopPrice:     o.StopPrice,
		TrailingDelta: o.TrailingDelta,
	}
}

// BaseQuantity is a utility function that will calculate the base quantity in case only quote quantity was provided
func (o *OrderRequest) BaseQuantity() float64 {
	if o.BaseQty != 0 {
		return o.BaseQty
	}

	return o.QuoteQty / o.Price
}

// QuoteQuantity is a utility function that will calculate the base quantity in case only quote quantity was provided
func (o *OrderRequest) QuoteQuantity() float64 {
	if o.QuoteQty != 0 {
		return o.QuoteQty
	}

	return o.BaseQty * o.Price
}

// FollowedOrderRequest is used to create an order and then follow it on a given plot.
type FollowedOrderRequest struct {
	StartAt, StopAt time.Time
	Symbol          string
	Interval        time.Duration
	Side            string
	BaseQuantity    float64
	QuoteQuantity   float64
	Plot            geometry.Plot
}

// FollowerOrder aggregates Follow data and Order data
type FollowedOrder struct {
	FollowID      string
	FollowRequest *FollowedOrderRequest
	Order         *Order
	follower      *geometry.Follower
}

// OrderedFollower is responsible for interacting with exchanges to manage orders
// and for following those orders.
type OrderFollower struct {
	Exchange       Exchange
	FollowedOrders []*FollowedOrder
}

// CreateFollowedOrder creates an order on given exchange and then follows it
func (of *OrderFollower) CreateFollowedOrder(foReq *FollowedOrderRequest) (*FollowedOrder, error) {
	follower := geometry.NewPlotFollower(foReq.Plot, foReq.Interval)

	order, err := of.Exchange.CreateOrder(OrderRequest{
		Symbol:   foReq.Symbol,
		Side:     foReq.Side,
		BaseQty:  foReq.BaseQuantity,
		QuoteQty: foReq.QuoteQuantity,
		Price:    follower.PriceAt(time.Now()),
	})
	if err != nil {
		return nil, err
	}

	if err := follower.Start(); err != nil {
		return nil, fmt.Errorf("unable to start follower: %w", err)
	}

	fo := &FollowedOrder{
		FollowID:      uuid.NewString(),
		FollowRequest: foReq,
		Order:         order,
		follower:      follower,
	}

	of.FollowedOrders = append(of.FollowedOrders, fo)

	go func() {
		err := of.followOrder(fo.Order, follower.TickerC())
		if err != nil {
			follower.Stop()
			logger.Errorf("error while following order: %w", err)
		}
	}()

	return fo, nil
}

// FollowOrder starts following an order that was already placed
func (of *OrderFollower) FollowOrder(orderID any, plot geometry.Plot, itv time.Duration) (*FollowedOrder, error) {
	order, err := of.Exchange.GetOrder(orderID)
	if err != nil {
		return nil, fmt.Errorf("error getting order %v from exchange: %w", orderID, err)
	}

	follower := geometry.NewPlotFollower(plot, itv)
	if err := follower.Start(); err != nil {
		return nil, fmt.Errorf("unable to start follower: %w", err)
	}

	fo := &FollowedOrder{
		FollowID: uuid.NewString(),
		Order:    order,
		follower: follower,
	}

	of.FollowedOrders = append(of.FollowedOrders, fo)

	go func() {
		if err := of.followOrder(fo.Order, follower.TickerC()); err != nil {
			follower.Stop()
		}
	}()

	return fo, nil
}

// CancelFollow stops following an order, if cancelOrder is true then it also cancels the order on the exchange
func (of *OrderFollower) CancelFollow(followID string, cancelOrder bool) error {
	//mutex
	for i, fo := range of.FollowedOrders {
		if fo.FollowID == followID {
			fo.follower.Stop()

			of.FollowedOrders = append(of.FollowedOrders[:i], of.FollowedOrders[i+1:]...)

			if cancelOrder {
				return of.Exchange.CancelOrder(fo.Order.ID)
			}
		}
	}

	return nil
}

func (of *OrderFollower) followOrder(order *Order, updateC chan geometry.FollowUpdate) error {
	for update := range updateC {
		pending, err := of.Exchange.OrderPending(order.ID)
		if err != nil {
			logger.Errorf("unable to check if order %v is pending: %w", order.ID, err)
		}

		if !pending {
			return fmt.Errorf("order %v not pending", order.ID)
		}

		if err := of.Exchange.CancelOrder(order.ID); err != nil {
			return fmt.Errorf("unable to cancel order %v: %w", order.ID, err)
		}

		newOrder, err := of.Exchange.CreateOrder(OrderRequest{
			Symbol:   order.Symbol,
			Side:     order.Side,
			BaseQty:  order.BaseQty,
			QuoteQty: order.QuoteQty,
			Price:    update.Price,
		})
		if err != nil {
			return fmt.Errorf("unable to create order: %w", err)
		}

		*order = *newOrder
	}

	return nil
}
