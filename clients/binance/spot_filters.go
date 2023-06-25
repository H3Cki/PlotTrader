package binance

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	binanceSDK "github.com/adshao/go-binance/v2"
)

var spotOrderTypeFilters = map[binanceSDK.OrderType]func(binanceSDK.Symbol, *SpotOrderRequest) error{
	binanceSDK.OrderTypeLimit: func(s binanceSDK.Symbol, or *SpotOrderRequest) error {
		// PRICE
		if pf := s.PriceFilter(); pf != nil {
			price, err := spotPriceFilter(pf, or.price)
			if err != nil {
				return err
			}

			or.price = price
		}

		// LOT SIZE
		if lsf := s.LotSizeFilter(); lsf != nil {
			qty, err := spotLotSizeFilter(lsf, or.BaseQuantity)
			if err != nil {
				return err
			}

			or.BaseQuantity = qty
		}

		// MIN NOTIONAL
		if mnf := s.MinNotionalFilter(); mnf != nil {
			err := spotMinNotionalFilter(mnf, or.price, or.BaseQuantity)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func applySpotFilters(s binanceSDK.Symbol, or *SpotOrderRequest) error {
	or.BaseQuantity = baseQuantity(or.price, or.BaseQuantity, or.QuoteQuantity)
	or.QuoteQuantity = quoteQuantity(or.price, or.BaseQuantity, or.QuoteQuantity)

	filterFunc, ok := spotOrderTypeFilters[binanceSDK.OrderType(or.OrderType)]
	if !ok {
		return fmt.Errorf("unsupported order type: %v", or.OrderType)
	}

	return filterFunc(s, or)
}

// priceFilter returns a price adjusted for the tickSize for a given symbol,
// returns an error if the price exceeds min or max value.
func spotPriceFilter(pf *binanceSDK.PriceFilter, price float64) (float64, error) {
	tickSize, err := strconv.ParseFloat(pf.TickSize, 64)
	if err != nil {
		return 0, err
	}

	newPrice := price

	if tickSize != 0 {
		// set price to nearest multiple of tickSize
		decimals := stringDecimalPlacesExp(pf.TickSize)
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

func spotLotSizeFilter(lsf *binanceSDK.LotSizeFilter, qty float64) (float64, error) {
	stepSize, err := strconv.ParseFloat(lsf.StepSize, 64)
	if err != nil {
		return 0, err
	}

	decimals := stringDecimalPlacesExp(lsf.StepSize)
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

func spotMinNotionalFilter(mnf *binanceSDK.NotionalFilter, price, qty float64) error {
	if mnf.MinNotional == "" {
		return nil
	}

	minNotional, err := strconv.ParseFloat(mnf.MinNotional, 64)
	if err != nil {
		return err
	}

	if price*qty < minNotional {
		return fmt.Errorf("minNotional too small, expected > %f, got %f", minNotional, price*qty)
	}

	return nil
}
