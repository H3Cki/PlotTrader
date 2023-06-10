package futures

import (
	"PlotTrader/trade/binance"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/H3Cki/TrendTrader/trade"
	futuresSDK "github.com/adshao/go-binance/v2/futures"
)

func applyFilters(s futuresSDK.Symbol, o *trade.Order) error {
	switch o.Type {
	case "limit":
		// PRICE
		if pf := s.PriceFilter(); pf != nil {
			price, err := priceFilter(pf, o.Price)
			if err != nil {
				return err
			}

			o.Price = price
		}

		// LOT SIZE
		if lsf := s.LotSizeFilter(); lsf != nil {
			qty, err := lotSizeFilter(lsf, o.BaseQty)
			if err != nil {
				return err
			}

			o.BaseQty = qty
		}

		// MIN NOTIONAL
		if mnf := s.MinNotionalFilter(); mnf != nil {
			err := minNotionalFilter(mnf, o.Price, o.BaseQty)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("usupported order type %s", o.Type)
	}

	return nil
}

func priceFilter(pf *futuresSDK.PriceFilter, price float64) (float64, error) {
	tickSize, err := strconv.ParseFloat(pf.TickSize, 64)
	if err != nil {
		return 0, err
	}

	newPrice := price

	if tickSize != 0 {
		// set price to nearest multiple of tickSize
		decimals := binance.StringDecimalPlacesExp(pf.TickSize)
		newPrice = math.Round(price/tickSize) * tickSize
		newPrice = math.Round(newPrice*decimals) / decimals
	}

	minPrice, err := strconv.ParseFloat(pf.MinPrice, 64)
	if err != nil {
		return 0, err
	}

	// reject if price is lower than min price
	if minPrice != 0 && newPrice < minPrice {
		return 0, nil
	}

	// reject is price is higher than max price
	maxPrice, err := strconv.ParseFloat(pf.MaxPrice, 64)
	if err != nil {
		return 0, err
	}

	if maxPrice != 0 && newPrice > maxPrice {
		return 0, nil
	}

	return newPrice, nil
}

func lotSizeFilter(lsf *futuresSDK.LotSizeFilter, qty float64) (float64, error) {
	stepSize, err := strconv.ParseFloat(lsf.StepSize, 64)
	if err != nil {
		return 0, err
	}

	decimals := binance.StringDecimalPlacesExp(lsf.StepSize)
	newQty := math.Floor(qty/stepSize) * stepSize
	newQty = math.Round(newQty*decimals) / decimals

	minQty, err := strconv.ParseFloat(lsf.MinQuantity, 64)
	if err != nil {
		return 0, err
	}

	if newQty < minQty {
		return 0, errors.New("quantity too small")
	}

	maxQty, err := strconv.ParseFloat(lsf.MaxQuantity, 64)
	if err != nil {
		return 0, err
	}

	if newQty > maxQty {
		return 0, errors.New("quantity too large")
	}

	return newQty, nil
}

func minNotionalFilter(mnf *futuresSDK.MinNotionalFilter, price, qty float64) error {
	if mnf.Notional == "" {
		return nil
	}

	minNotional, err := strconv.ParseFloat(mnf.Notional, 64)
	if err != nil {
		return err
	}

	if price*qty < minNotional {
		return fmt.Errorf("minNotional too small, expected > %f, got %f", minNotional, price*qty)
	}

	return nil
}
