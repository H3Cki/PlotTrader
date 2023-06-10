package futures

import (
	"PlotTrader/trade/binance"
	"encoding/json"
	"math"
	"testing"

	"github.com/H3Cki/TrendTrader/trade"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/stretchr/testify/assert"
)

var ethbtc = `
    {
		"symbol": "BNBBTC",
		"status": "TRADING",
		"baseAsset": "BNB",
		"baseAssetPrecision": 8,
		"quoteAsset": "BTC",
		"quotePrecision": 8,
		"quoteAssetPrecision": 8,
		"baseCommissionPrecision": 8,
		"quoteCommissionPrecision": 8,
		"Types": [
		  "LIMIT",
		  "LIMIT_MAKER",
		  "MARKET",
		  "STOP_LOSS_LIMIT",
		  "TAKE_PROFIT_LIMIT"
		],
		"icebergAllowed": true,
		"ocoAllowed": true,
		"quoteOrderQtyMarketAllowed": true,
		"allowTrailingStop": true,
		"cancelReplaceAllowed": true,
		"isSpotTradingAllowed": true,
		"isMarginTradingAllowed": false,
		"filters": [
		  {
			"filterType": "PRICE_FILTER",
			"minPrice": "0.00000100",
			"maxPrice": "10.00000000",
			"tickSize": "0.00000100"
		  },
		  {
			"filterType": "PERCENT_PRICE",
			"multiplierUp": "5",
			"multiplierDown": "0.2",
			"avgPriceMins": 1
		  },
		  {
			"filterType": "LOT_SIZE",
			"minQty": "0.01000000",
			"maxQty": "9000.00000000",
			"stepSize": "0.01000000"
		  },
		  {
			"filterType": "MIN_NOTIONAL",
			"notional": "0.00010000",
			"applyToMarket": true,
			"avgPriceMins": 1
		  },
		  { "filterType": "ICEBERG_PARTS", "limit": 10 },
		  {
			"filterType": "MARKET_LOT_SIZE",
			"minQty": "0.00000000",
			"maxQty": "1000.00000000",
			"stepSize": "0.00000000"
		  },
		  {
			"filterType": "TRAILING_DELTA",
			"minTrailingAboveDelta": 10,
			"maxTrailingAboveDelta": 2000,
			"minTrailingBelowDelta": 10,
			"maxTrailingBelowDelta": 2000
		  },
		  { "filterType": "MAX_NUM_ORDERS", "maxNumOrders": 200 },
		  { "filterType": "MAX_NUM_ALGO_ORDERS", "maxNumAlgoOrders": 5 }
		],
		"permissions": ["SPOT"]
    }`

func Test_applyFilters(t *testing.T) {
	var symbolETHBTC futures.Symbol

	if err := json.Unmarshal([]byte(ethbtc), &symbolETHBTC); err != nil {
		t.Error(err)
	}

	type args struct {
		s futures.Symbol
		o trade.Order
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		exRes   trade.Order
	}{
		{
			name: "1",
			args: args{
				s: symbolETHBTC,
				o: trade.Order{
					Type:    "limit",
					Symbol:  "ETHBTC",
					Price:   0.12345678912345,
					BaseQty: 0.212345678912345,
				},
			},
			wantErr: false,
			exRes: trade.Order{
				Type:    "limit",
				Symbol:  "ETHBTC",
				Price:   0.123457,
				BaseQty: 0.21,
			},
		},
		{
			name: "quantity too small",
			args: args{
				s: symbolETHBTC,
				o: trade.Order{
					Type:    "limit",
					Symbol:  "ETHBTC",
					Price:   0.078794,
					BaseQty: 0.0001,
				},
			},
			wantErr: true,
			exRes: trade.Order{
				Type:    "limit",
				Symbol:  "ETHBTC",
				Price:   0.078794,
				BaseQty: 0.0001,
			},
		},
		{
			name: "quantity too big",
			args: args{
				s: symbolETHBTC,
				o: trade.Order{
					Type:    "limit",
					Symbol:  "ETHBTC",
					Price:   0.078794,
					BaseQty: 0.0001,
				},
			},
			wantErr: true,
			exRes: trade.Order{
				Type:    "limit",
				Symbol:  "ETHBTC",
				Price:   0.078794,
				BaseQty: 100000.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applyFilters(tt.args.s, &tt.args.o)

			assert.Equal(t, (err != nil), tt.wantErr)
			if tt.wantErr {
				return
			}
			assert.Equal(t, tt.exRes.Price, tt.args.o.Price)
			assert.LessOrEqual(t, math.Abs(binance.Gain(tt.exRes.BaseQty, tt.args.o.BaseQty)), 0.001)
		})
	}
}

func Test_priceFilter(t *testing.T) {
	var symbolETHBTC futures.Symbol

	if err := json.Unmarshal([]byte(ethbtc), &symbolETHBTC); err != nil {
		t.Error(err)
	}

	pf := symbolETHBTC.PriceFilter()

	tests := []struct {
		name    string
		price   float64
		want    float64
		wantErr bool
	}{
		{
			"1",
			1.1111111111, // 10
			1.111111,
			false,
		},
		{
			"2",
			1.1111119111, // 10
			1.111112,
			false,
		},
		{
			"3",
			1.000000000000000000000000000000000000000009, // 10
			1.0,
			false,
		},
		{
			"minPrice",
			0.000000000000000000000000000000000000000009, // 10
			0,
			false,
		},
		{
			"maxPrice",
			11.000000000000000000000000000000000000000009, // 10
			0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := priceFilter(pf, tt.price)
			if (err != nil) != tt.wantErr {
				t.Errorf("priceFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("priceFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_lotSizeFilter(t *testing.T) {
	var symbolETHBTC futures.Symbol

	if err := json.Unmarshal([]byte(ethbtc), &symbolETHBTC); err != nil {
		t.Error(err)
	}

	pf := symbolETHBTC.LotSizeFilter()

	tests := []struct {
		name    string
		qty     float64
		want    float64
		wantErr bool
	}{
		{
			"1",
			123.1111111111, // 10
			123.11,
			false,
		},
		{
			"2",
			123.1259111119111, // 10
			123.12,
			false,
		},
		{
			"3",
			123.000000000000000000000000000000000000000009, // 10
			123.0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lotSizeFilter(pf, tt.qty)
			if (err != nil) != tt.wantErr {
				t.Errorf("lotSizeFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("lotSizeFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
