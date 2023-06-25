package binance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/H3Cki/Plotor/logger"
	"github.com/H3Cki/Plotor/plotor"
	"github.com/google/uuid"

	"github.com/adshao/go-binance/v2"
	sdk "github.com/adshao/go-binance/v2"
)

func init() {
	sdk.UseTestnet = true
}

var SPOT_EXCHANGEINFO_FILENAME = "spot_exchange_info.json"

// SpotOrderRequest holds fields that are required (or supported) to create an order
type SpotOrderRequest struct {
	Symbol        string              `json:"symbol"`
	Side          sdk.SideType        `json:"side"`
	OrderType     sdk.OrderType       `json:"type"`
	TimeInForce   sdk.TimeInForceType `json:"timeInForce"`
	BaseQuantity  float64             `json:"baseQuantity"`
	QuoteQuantity float64             `json:"quoteQuentity"`
	ClientOrderID string              `json:"clientOrderID"`
	price         float64             //internal use
}

type SpotOrder struct {
	Symbol                   string                  `json:"symbol"`
	OrderID                  int64                   `json:"orderId"`
	OrderListId              int64                   `json:"orderListId"`
	ClientOrderID            string                  `json:"clientOrderId"`
	Price                    string                  `json:"price"`
	OrigQuantity             string                  `json:"origQty"`
	ExecutedQuantity         string                  `json:"executedQty"`
	CummulativeQuoteQuantity string                  `json:"cummulativeQuoteQty"`
	Status                   binance.OrderStatusType `json:"status"`
	TimeInForce              binance.TimeInForceType `json:"timeInForce"`
	Type                     binance.OrderType       `json:"type"`
	Side                     binance.SideType        `json:"side"`
	//StopPrice                string                  `json:"stopPrice"`
	//IcebergQuantity          string                  `json:"icebergQty"`
	Time                   int64  `json:"time"`
	UpdateTime             int64  `json:"updateTime"`
	IsWorking              bool   `json:"isWorking"`
	IsIsolated             bool   `json:"isIsolated"`
	OrigQuoteOrderQuantity string `json:"origQuoteOrderQty"`
}

func (o *SpotOrder) Details() (map[string]any, error) {
	m := map[string]any{}

	bytes, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}

	return m, nil
}

type SpotCredentials struct {
	API_KEY, SECRET_KEY string
}

type SpotClient struct {
	sdkClient *sdk.Client
	ei        sdk.ExchangeInfo
}

func (s *SpotClient) SetUp(creds SpotCredentials) error {
	s.sdkClient = binance.NewClient(creds.API_KEY, creds.SECRET_KEY)

	err := s.exchangeInfoFromFile()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Errorf("error loading exchange info from file: %v", err)
	}

	if err == nil && !eiOutdated(s.ei) {
		return nil
	}

	if err := s.exchangeInfo(); err != nil {
		return fmt.Errorf("error loading exchange info: %w", err)
	}

	return nil
}

func (s *SpotClient) GetOrder(ctx context.Context, order plotor.ClientOrder) (plotor.ClientOrder, error) {
	return s.getOrder(ctx, order)
}

// GetOrder fetches current order state, requires OrderID and Symbol to be set
func (s *SpotClient) getOrder(ctx context.Context, order plotor.ClientOrder) (*SpotOrder, error) {
	req, ok := order.(*SpotOrder)
	if !ok {
		return nil, fmt.Errorf("unexpected order type: %v", order)
	}

	res, err := s.sdkClient.NewGetOrderService().OrderID(req.OrderID).Symbol(req.Symbol).Do(ctx)
	if err != nil {
		return nil, err
	}

	return &SpotOrder{
		Symbol:                   res.Symbol,
		OrderID:                  res.OrderID,
		OrderListId:              res.OrderListId,
		ClientOrderID:            res.ClientOrderID,
		Price:                    res.Price,
		OrigQuantity:             res.OrigQuantity,
		ExecutedQuantity:         res.ExecutedQuantity,
		CummulativeQuoteQuantity: res.CummulativeQuoteQuantity,
		Status:                   res.Status,
		TimeInForce:              res.TimeInForce,
		Type:                     res.Type,
		Side:                     res.Side,
		//StopPrice:                res.StopPrice,
		//IcebergQuantity:          res.IcebergQuantity,
		Time:                   res.Time,
		UpdateTime:             res.UpdateTime,
		IsWorking:              res.IsWorking,
		IsIsolated:             res.IsIsolated,
		OrigQuoteOrderQuantity: res.OrigQuoteOrderQuantity,
	}, nil
}

func (e *SpotClient) CreateOrder(ctx context.Context, orderData any, price float64) (plotor.ClientOrder, error) {
	req := &SpotOrderRequest{}

	switch v := orderData.(type) {
	case *SpotOrderRequest:
		req = v
	case SpotOrderRequest:
		req = &v
	case []byte:
		if err := json.Unmarshal(v, req); err != nil {
			return nil, fmt.Errorf("unable to unmarshal order data: %w", err)
		}
	case json.RawMessage:
		if err := json.Unmarshal(v, req); err != nil {
			return nil, fmt.Errorf("unable to unmarshal order data: %w", err)
		}
	default:
		return nil, fmt.Errorf("unexpected order data type: %v", v)
	}

	req.price = price
	if req.ClientOrderID == "" {
		req.ClientOrderID = uuid.NewString()
	}

	order, err := e.createOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	return e.GetOrder(ctx, order)
}

func (e *SpotClient) createOrder(ctx context.Context, req *SpotOrderRequest) (*SpotOrder, error) {
	exchangeSymbol, err := e.symbol(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	if err := applySpotFilters(exchangeSymbol, req); err != nil {
		return nil, fmt.Errorf("error filtering order request: %w", err)
	}

	orderSvc := e.sdkClient.NewCreateOrderService()

	orderSvc.NewClientOrderID(req.ClientOrderID).
		Side(sdk.SideType(req.Side)).
		Side(sdk.SideType(req.Side)).
		Type(sdk.OrderType(req.OrderType)).
		Symbol(req.Symbol).
		Price(fmt.Sprint(req.price)).
		Quantity(fmt.Sprint(baseQuantity(req.price, req.BaseQuantity, req.QuoteQuantity))).
		TimeInForce(sdk.TimeInForceType(req.TimeInForce))

	res, err := orderSvc.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	return &SpotOrder{
		Symbol:                   res.Symbol,
		OrderID:                  res.OrderID,
		ClientOrderID:            res.ClientOrderID,
		Price:                    res.Price,
		OrigQuantity:             res.OrigQuantity,
		ExecutedQuantity:         res.ExecutedQuantity,
		CummulativeQuoteQuantity: res.CummulativeQuoteQuantity,
		Status:                   res.Status,
		TimeInForce:              res.TimeInForce,
		Type:                     res.Type,
		Side:                     res.Side,
	}, nil
}

func (e *SpotClient) UpdateOrderPrice(ctx context.Context, order plotor.ClientOrder, price float64) (plotor.ClientOrder, error) {
	o, err := e.getOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	if err := e.CancelOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("error cancelling order: %w", err)
	}

	origBaseQty, err := strconv.ParseFloat(o.OrigQuantity, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing OrigQuantity: %w", err)
	}

	execBaseQty, err := strconv.ParseFloat(o.ExecutedQuantity, 64)
	if err != nil {
		return nil, fmt.Errorf("error parsing OrigQuantity: %w", err)
	}

	or := &SpotOrderRequest{
		ClientOrderID: o.ClientOrderID,
		Symbol:        o.Symbol,
		Side:          o.Side,
		OrderType:     o.Type,
		TimeInForce:   o.TimeInForce,
		BaseQuantity:  origBaseQty - execBaseQty,
		price:         price,
	}

	res, err := e.createOrder(ctx, or)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (e *SpotClient) CancelOrder(ctx context.Context, order plotor.ClientOrder) error {
	o, ok := order.(*SpotOrder)
	if !ok {
		return fmt.Errorf("unexpected order type: %v", order)
	}

	_, err := e.sdkClient.NewCancelOrderService().OrderID(o.OrderID).Symbol(o.Symbol).Do(ctx)

	return err
}

func (c *SpotClient) symbol(ctx context.Context, symbol string) (sdk.Symbol, error) {
	fetched := false

	if eiOutdated(c.ei) {
		if err := c.exchangeInfo(); err != nil {
			logger.Errorf("error updating exchange info: %v", err)
		} else {
			fetched = true
		}
	}

	for _, fsymbol := range c.ei.Symbols {
		if fsymbol.Symbol == symbol {
			return fsymbol, nil
		}
	}

	// second chance, ei was loaded form file but the symbol might be new and require a reload
	if !fetched {
		if err := c.exchangeInfo(); err != nil {
			return sdk.Symbol{}, err
		}

		for _, fsymbol := range c.ei.Symbols {
			if fsymbol.Symbol == symbol {
				return fsymbol, nil
			}
		}
	}

	return sdk.Symbol{}, fmt.Errorf("unknown symbol: %s", symbol)
}

func (c *SpotClient) exchangeInfo() error {
	res, err := c.sdkClient.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return fmt.Errorf("unable to fetch spot exchange info: %w", err)
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("unable to marshal exchange info: %w", err)
	}

	err = os.WriteFile(SPOT_EXCHANGEINFO_FILENAME, bytes, 0o777)
	if err != nil {
		logger.Errorf("\nunable to save exchange info to file: %w", err)
	}

	c.ei = *res

	return nil
}

func (s *SpotClient) exchangeInfoFromFile() error {
	bytes, err := os.ReadFile(SPOT_EXCHANGEINFO_FILENAME)
	if err != nil {
		return err
	}

	ei := &sdk.ExchangeInfo{}
	if err := json.Unmarshal(bytes, ei); err != nil {
		return err
	}

	s.ei = *ei

	return nil
}

func eiOutdated(ei sdk.ExchangeInfo) bool {
	return time.Since(time.Unix(ei.ServerTime, 0)) > time.Hour*24
}
