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
	"github.com/adshao/go-binance/v2/futures"
	"github.com/google/uuid"
)

func init() {
	futures.UseTestnet = true
}

var FUTURES_EXCHANGEINFO_FILENAME = "futures_exchange_info.json"

// FuturesOrderRequest holds fields that are required (or supported) to create an order
type FuturesOrderRequest struct {
	Symbol        string                  `json:"symbol"`
	Side          futures.SideType        `json:"side"`
	OrderType     futures.OrderType       `json:"type"`
	TimeInForce   futures.TimeInForceType `json:"timeInForce"`
	BaseQuantity  float64                 `json:"baseQuantity"`
	QuoteQuantity float64                 `json:"quoteQuentity"`
	ClientOrderID string                  `json:"clientOrderID"`
	price         float64                 //internal use
}

type FuturesOrder struct {
	Symbol           string                   `json:"symbol"`
	OrderID          int64                    `json:"orderId"`
	ClientOrderID    string                   `json:"clientOrderId"`
	Price            string                   `json:"price"`
	ReduceOnly       bool                     `json:"reduceOnly"`
	OrigQuantity     string                   `json:"origQty"`
	ExecutedQuantity string                   `json:"executedQty"`
	CumQuantity      string                   `json:"cumQty"`
	CumQuote         string                   `json:"cumQuote"`
	Status           futures.OrderStatusType  `json:"status"`
	TimeInForce      futures.TimeInForceType  `json:"timeInForce"`
	Type             futures.OrderType        `json:"type"`
	Side             futures.SideType         `json:"side"`
	StopPrice        string                   `json:"stopPrice"`
	Time             int64                    `json:"time"`
	UpdateTime       int64                    `json:"updateTime"`
	WorkingType      futures.WorkingType      `json:"workingType"`
	ActivatePrice    string                   `json:"activatePrice"`
	PriceRate        string                   `json:"priceRate"`
	AvgPrice         string                   `json:"avgPrice"`
	OrigType         string                   `json:"origType"`
	PositionSide     futures.PositionSideType `json:"positionSide"`
	PriceProtect     bool                     `json:"priceProtect"`
	ClosePosition    bool                     `json:"closePosition"`
}

func (o *FuturesOrder) Details() (map[string]any, error) {
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

type FuturesCredentials struct {
	API_KEY, SECRET_KEY string
}

type FuturesClient struct {
	sdkClient *futures.Client
	ei        futures.ExchangeInfo
}

func (s *FuturesClient) SetUp(creds FuturesCredentials) error {
	s.sdkClient = futures.NewClient(creds.API_KEY, creds.SECRET_KEY)

	err := s.exchangeInfoFromFile()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		logger.Errorf("error loading exchange info from file: %v", err)
	}

	if err == nil && s.eiOutdated() {
		return nil
	}

	if err := s.exchangeInfo(); err != nil {
		return fmt.Errorf("error loading exchange info: %w", err)
	}

	return nil
}

func (s *FuturesClient) GetOrder(ctx context.Context, order plotor.ClientOrder) (plotor.ClientOrder, error) {
	return s.getOrder(ctx, order)
}

// GetOrder fetches current order state, requires OrderID and Symbol to be set
func (s *FuturesClient) getOrder(ctx context.Context, order plotor.ClientOrder) (*FuturesOrder, error) {
	req, ok := order.(*FuturesOrder)
	if !ok {
		return nil, fmt.Errorf("unexpected order type: %v", order)
	}

	res, err := s.sdkClient.NewGetOrderService().OrderID(req.OrderID).Symbol(req.Symbol).Do(ctx)
	if err != nil {
		return nil, err
	}

	return &FuturesOrder{
		Symbol:           res.Symbol,
		OrderID:          res.OrderID,
		ClientOrderID:    res.ClientOrderID,
		Price:            res.Price,
		ReduceOnly:       res.ReduceOnly,
		OrigQuantity:     res.OrigQuantity,
		ExecutedQuantity: res.ExecutedQuantity,
		CumQuantity:      res.CumQuantity,
		CumQuote:         res.CumQuote,
		Status:           res.Status,
		TimeInForce:      res.TimeInForce,
		Type:             res.Type,
		Side:             res.Side,
		StopPrice:        res.StopPrice,
		Time:             res.Time,
		UpdateTime:       res.UpdateTime,
		WorkingType:      res.WorkingType,
		ActivatePrice:    res.ActivatePrice,
		PriceRate:        res.PriceRate,
		AvgPrice:         res.AvgPrice,
		OrigType:         res.OrigType,
		PositionSide:     res.PositionSide,
		PriceProtect:     res.PriceProtect,
		ClosePosition:    res.ClosePosition,
	}, nil
}

func (e *FuturesClient) CreateOrder(ctx context.Context, orderData any, price float64) (plotor.ClientOrder, error) {
	req := &FuturesOrderRequest{}

	switch v := orderData.(type) {
	case *FuturesOrderRequest:
		req = v
	case FuturesOrderRequest:
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

func (e *FuturesClient) createOrder(ctx context.Context, req *FuturesOrderRequest) (*FuturesOrder, error) {
	exchangeSymbol, err := e.symbol(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	if err := applyFuturesFilters(exchangeSymbol, req); err != nil {
		return nil, fmt.Errorf("error filtering order request: %w", err)
	}

	orderSvc := e.sdkClient.NewCreateOrderService()

	orderSvc.NewClientOrderID(req.ClientOrderID).
		Side(futures.SideType(req.Side)).
		Side(futures.SideType(req.Side)).
		Type(futures.OrderType(req.OrderType)).
		Symbol(req.Symbol).
		Price(fmt.Sprint(req.price)).
		Quantity(fmt.Sprint(baseQuantity(req.price, req.BaseQuantity, req.QuoteQuantity))).
		TimeInForce(futures.TimeInForceType(req.TimeInForce))

	res, err := orderSvc.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}

	return &FuturesOrder{
		Symbol:           res.Symbol,
		OrderID:          res.OrderID,
		ClientOrderID:    res.ClientOrderID,
		Price:            res.Price,
		ReduceOnly:       res.ReduceOnly,
		OrigQuantity:     res.OrigQuantity,
		ExecutedQuantity: res.ExecutedQuantity,
		// CumQuantity:      "",
		CumQuote:    res.CumQuote,
		Status:      res.Status,
		TimeInForce: res.TimeInForce,
		Type:        res.Type,
		Side:        res.Side,
		StopPrice:   res.StopPrice,
		//Time:             0,
		UpdateTime:    res.UpdateTime,
		WorkingType:   res.WorkingType,
		ActivatePrice: res.ActivatePrice,
		PriceRate:     res.PriceRate,
		AvgPrice:      res.AvgPrice,
		//OrigType:      "",
		PositionSide:  res.PositionSide,
		PriceProtect:  res.PriceProtect,
		ClosePosition: res.ClosePosition,
	}, nil
}

func (f *FuturesClient) UpdateOrderPrice(ctx context.Context, order plotor.ClientOrder, price float64) (plotor.ClientOrder, error) {
	o, err := f.getOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	if err := f.CancelOrder(ctx, order); err != nil {
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

	or := &FuturesOrderRequest{
		ClientOrderID: o.ClientOrderID,
		Symbol:        o.Symbol,
		Side:          o.Side,
		OrderType:     o.Type,
		TimeInForce:   o.TimeInForce,
		BaseQuantity:  origBaseQty - execBaseQty,
		price:         price,
	}

	res, err := f.createOrder(ctx, or)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (e *FuturesClient) CancelOrder(ctx context.Context, order plotor.ClientOrder) error {
	o, ok := order.(*FuturesOrder)
	if !ok {
		return fmt.Errorf("unexpected order type: %v", order)
	}

	_, err := e.sdkClient.NewCancelOrderService().OrderID(o.OrderID).Symbol(o.Symbol).Do(ctx)

	return err
}

func (f *FuturesClient) symbol(ctx context.Context, symbol string) (futures.Symbol, error) {
	fetched := false

	if f.eiOutdated() {
		if err := f.exchangeInfo(); err != nil {
			logger.Errorf("error updating exchange info: %v", err)
		} else {
			fetched = true
		}
	}

	for _, fsymbol := range f.ei.Symbols {
		if fsymbol.Symbol == symbol {
			return fsymbol, nil
		}
	}

	// second chance, ei was loaded form file but the symbol might be new and require a reload
	if !fetched {
		if err := f.exchangeInfo(); err != nil {
			return futures.Symbol{}, err
		}

		for _, fsymbol := range f.ei.Symbols {
			if fsymbol.Symbol == symbol {
				return fsymbol, nil
			}
		}
	}

	return futures.Symbol{}, fmt.Errorf("unknown symbol: %s", symbol)
}

func (f *FuturesClient) exchangeInfo() error {
	res, err := f.sdkClient.NewExchangeInfoService().Do(context.Background())
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

	f.ei = *res

	return nil
}

func (f *FuturesClient) exchangeInfoFromFile() error {
	bytes, err := os.ReadFile(SPOT_EXCHANGEINFO_FILENAME)
	if err != nil {
		return err
	}

	ei := &futures.ExchangeInfo{}
	if err := json.Unmarshal(bytes, ei); err != nil {
		return err
	}

	f.ei = *ei

	return nil
}

func (f *FuturesClient) eiOutdated() bool {
	return time.Since(time.Unix(f.ei.ServerTime, 0)) > time.Hour*24
}
