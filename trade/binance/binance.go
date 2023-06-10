package binance

import (
	"github.com/H3Cki/TrendTrader/trade"
)

var id = 0

type Exchange struct {
}

func (e *Exchange) Name() string {
	return "binance"
}

func (e *Exchange) OrderPending(orderID any) (bool, error) {
	return true, nil
}

func (e *Exchange) CancelOrder(orderID any) error {
	return nil
}

func (e *Exchange) CreateOrder(orderReq *trade.OrderRequest) (o *trade.Order, err error) {
	id += 1
	o = orderReq.ToOrder()
	o.ID = id
	return o, nil
}

func (e *Exchange) GetOrder(any) (*trade.Order, error) {
	id += 1
	return &trade.Order{ID: id}, nil
}
