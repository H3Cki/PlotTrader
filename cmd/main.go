package main

import (
	"github.com/H3Cki/PlotTrader/trade"
	"github.com/H3Cki/PlotTrader/trade/binance"
)

func main() {
	ex := binance.NewClient(binance.Credentials{})
	fol := trade.OrderFollower{
		Exchange: ex,
	}

	fol.CreateFollowedOrder(&trade.FollowedOrderRequest{})
}
