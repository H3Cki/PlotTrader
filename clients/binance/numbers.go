package binance

import (
	"math"
	"strings"
)

func stringDecimalPlaces(s string) int {
	s = strings.Trim(s, "0")
	i := strings.IndexByte(s, '.')

	if i > -1 {
		return len(s) - i - 1
	}

	return 0
}

func stringDecimalPlacesExp(s string) float64 {
	n := stringDecimalPlaces(s)
	return decimalPlacesToExp(n)
}

func decimalPlacesToExp(n int) float64 {
	if n == 0 {
		return 1
	}

	return math.Pow(10, float64(n))
}

const prec = 1000000000

func gain(after, before float64) float64 {
	v := (after - before) / before
	return math.Round(v*prec) / prec
}

func baseQuantity(price, base, quote float64) float64 {
	if base != 0 {
		return base
	}

	return quote / price
}

func quoteQuantity(price, base, quote float64) float64 {
	if quote != 0 {
		return quote
	}

	return base * price
}
