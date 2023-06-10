package spot

import (
	"PlotTrader/logger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/H3Cki/TrendTrader/trade"
	binanceSDK "github.com/adshao/go-binance/v2"
)

var EXCHANGEINFO_FILENAME = "exchange_info.json"

type Credentials struct {
	API_KEY, SECRET_KEY string
}

type Client struct {
	sdkClient *binanceSDK.Client
	ei        *binanceSDK.ExchangeInfo
}

func NewClient(creds Credentials) *Client {
	return &Client{
		sdkClient: binanceSDK.NewClient(creds.API_KEY, creds.SECRET_KEY),
	}
}

func (e *Client) Name() string {
	return "binance"
}

func (e *Client) OrderPending(orderID any) (bool, error) {
	oid, ok := orderID.(int64)
	if !ok {
		return false, fmt.Errorf("unable to cast orderID %v to type int64", orderID)
	}

	res, err := e.sdkClient.NewGetOrderService().OrderID(oid).Do(context.Background())
	if err != nil {
		return false, err
	}

	return res.IsWorking, nil
}

func (e *Client) CancelOrder(orderID any) error {
	oid, ok := orderID.(int64)
	if !ok {
		return fmt.Errorf("unable to cast orderID %v to type int64", orderID)
	}

	_, err := e.sdkClient.NewCancelOrderService().OrderID(oid).Do(context.Background())

	return err
}

func (e *Client) CreateOrder(orderReq trade.OrderRequest) (*trade.Order, error) {
	ctx := context.Background()

	exchangeSymbol, err := e.symbol(ctx, orderReq.Symbol)
	if err != nil {
		return nil, err
	}

	if err := applyFilters(exchangeSymbol, &orderReq); err != nil {
		return nil, fmt.Errorf("error filtering order request: %w", err)
	}

	res, err := e.sdkClient.NewCreateOrderService().
		Side(binanceSDK.SideType(orderReq.Side)).
		Type(binanceSDK.OrderType(orderReq.Type)).
		Symbol(orderReq.Symbol).
		Price(fmt.Sprint(orderReq.Price)).
		StopPrice(fmt.Sprint(orderReq.StopPrice)).
		TrailingDelta(fmt.Sprint(orderReq.TrailingDelta)).
		Quantity(fmt.Sprint(orderReq.BaseQuantity())).
		TimeInForce(binanceSDK.TimeInForceType(orderReq.TimeInForce)).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	return &trade.Order{
		ID:            res.OrderID,
		Symbol:        res.Symbol,
		Side:          string(res.Side),
		Type:          string(res.Type),
		Price:         orderReq.Price,
		BaseQty:       orderReq.BaseQty,
		QuoteQty:      orderReq.QuoteQty,
		TimeInForce:   string(res.TimeInForce),
		StopPrice:     orderReq.StopPrice,
		TrailingDelta: orderReq.TrailingDelta,
	}, err
}

func (e *Client) GetOrder(orderID any) (*trade.Order, error) {
	oid, ok := orderID.(int64)
	if !ok {
		return nil, fmt.Errorf("unable to cast orderID %v to type int64", orderID)
	}

	res, err := e.sdkClient.NewGetOrderService().OrderID(oid).Do(context.Background())
	if err != nil {
		return nil, err
	}

	price, err := strconv.ParseFloat(res.Price, 64)
	if err != nil {
		return nil, err
	}

	baseQty, err := strconv.ParseFloat(res.OrigQuantity, 64)
	if err != nil {
		return nil, err
	}

	quoteQty, err := strconv.ParseFloat(res.CummulativeQuoteQuantity, 64)
	if err != nil {
		return nil, err
	}

	stopPrice, err := strconv.ParseFloat(res.StopPrice, 64)
	if err != nil {
		return nil, err
	}

	return &trade.Order{
		ID:            res.OrderID,
		Symbol:        res.Symbol,
		Side:          string(res.Side),
		Type:          string(res.Type),
		Price:         price,
		BaseQty:       baseQty,
		QuoteQty:      quoteQty,
		TimeInForce:   string(res.TimeInForce),
		StopPrice:     stopPrice,
		TrailingDelta: 0.0,
	}, nil
}

func (c *Client) symbol(ctx context.Context, symbol string) (binanceSDK.Symbol, error) {
	for _, fsymbol := range c.ei.Symbols {
		if fsymbol.Symbol == symbol {
			return fsymbol, nil
		}
	}

	err := c.exchangeInfo(true)
	if err != nil {
		return binanceSDK.Symbol{}, err
	}

	for _, fsymbol := range c.ei.Symbols {
		if fsymbol.Symbol == symbol {
			return fsymbol, nil
		}
	}

	return binanceSDK.Symbol{}, fmt.Errorf("unknown symbol %s", symbol)
}

func (c *Client) exchangeInfo(forceUpdate bool) error {
	if !forceUpdate {
		bytes, err := os.ReadFile(EXCHANGEINFO_FILENAME)
		if err == nil {
			ei := &binanceSDK.ExchangeInfo{}

			err = json.Unmarshal(bytes, ei)
			if err != nil {
				return err
			}

			c.ei = ei

			return nil
		}
	}

	res, err := c.sdkClient.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return fmt.Errorf("unable to fetch futures exchange info: %w", err)
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("unable to marshal exchange info: %w", err)
	}

	err = os.WriteFile(EXCHANGEINFO_FILENAME, bytes, 0o777)
	if err != nil {
		logger.Errorf("\nunable to save exchange info to file: %w", err)
	}

	c.ei = res

	return nil
}
