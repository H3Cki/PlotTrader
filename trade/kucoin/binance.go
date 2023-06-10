package binance

import (
	"github.com/H3Cki/PlotTrader/logger"
	"github.com/H3Cki/PlotTrader/trade"
)

var id = 0

type Exchange struct {
}

func (e *Exchange) Name() string {
	return "kucoin"
}

func (e *Exchange) OrderPending(orderID any) (bool, error) {
	return true, nil
}

func (e *Exchange) CancelOrder(orderID any) error {
	return nil
}

func (e *Exchange) CreateOrder(orderReq trade.OrderRequest) (orderID any, err error) {
	id += 1
	logger.Infof("created order id %d", id)
	return id, nil
}
