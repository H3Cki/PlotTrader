package main

import (
	"fmt"
	"time"

	"github.com/H3Cki/TrendTrader/geometry"
	"github.com/H3Cki/TrendTrader/logger"
	"github.com/H3Cki/TrendTrader/trade"
	"github.com/H3Cki/TrendTrader/trade/binance"
)

func main() {
	ex := &binance.Exchange{}
	pf := trade.NewPlotFollower(ex)

	fr := trade.FollowRequest{
		Symbol:        "BTCTEST",
		Interval:      "1s",
		Side:          "buy",
		BaseQuantity:  10.0,
		QuoteQuantity: 0.0,
		Plot:          geometry.NewOffsetPlot(geometry.NewLine(geometry.Point{}, geometry.Point{}), geometry.NewAbsoluteOffset(10)),
	}

	_, err := pf.Follow(fr)
	if err != nil {
		logger.Fatal(err)

	}

	for {
		time.Sleep(time.Second)
		fmt.Println(pf.Follows()[0].OrderID)
	}

	select {}
}
