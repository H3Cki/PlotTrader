package main

import (
	"fmt"
	"time"

	"github.com/H3Cki/PlotTrader/trade/binance/spot"
	"github.com/H3Cki/TrendTrader/geometry"
	"github.com/H3Cki/TrendTrader/trade"
	"github.com/H3Cki/TrendTrader/trade/binance"
)

func main() {
	trader := trade.OrderFollower{
		Exchange:       &binance.Exchange{},
		FollowedOrders: []*trade.FollowedOrder{},
	}

	fr := &trade.FollowedOrderRequest{
		Symbol:        "BTCTEST",
		Interval:      time.Second,
		Side:          "buy",
		BaseQuantity:  10.0,
		QuoteQuantity: 0.0,
		Plot:          geometry.NewOffsetPlot(geometry.NewLine(geometry.Point{Date: time.Now(), Price: 1}, geometry.Point{Date: time.Now().Add(10 * time.Second), Price: 10}), geometry.NewAbsoluteOffset(0)),
	}

	fo, err := trader.CreateFollowedOrder(fr)
	if err != nil {
		panic(err)
	}

	for {
		time.Sleep(fr.Interval + 1)

		fmt.Printf("%+v\n", fo.Order)
	}

	ex := spot.NewClient(spot.Credentials{})
	fol := trade.OrderFollower{
		Exchange: ex,
	}
}
