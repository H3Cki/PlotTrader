package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/H3Cki/Plotor/geometry"
	"github.com/H3Cki/Plotor/plotor"
	"github.com/gin-gonic/gin"
)

type createPlotOrderRequest struct {
	Interval string
	Plot     json.RawMessage
	Order    json.RawMessage
}

type createPlotOrderResponse struct {
	PlotOrderID string
	ClientOrder any
	Error       string
}

type getPlotOrderResponse struct {
	PlotOrderID string
	ClientOrder map[string]any
	Plot        map[string]any
	Interval    string
	LastTick    time.Time
	Error       string
}

func gpoErr(prefix string, err error) getPlotOrderResponse {
	if err != nil {
		return getPlotOrderResponse{
			Error: fmt.Sprintf("%s: %s", prefix, err.Error()),
		}
	}

	return getPlotOrderResponse{
		Error: prefix,
	}
}

func cpoErr(prefix string, err error) createPlotOrderResponse {
	if err != nil {
		return createPlotOrderResponse{
			Error: fmt.Sprintf("%s: %s", prefix, err.Error()),
		}
	}

	return createPlotOrderResponse{
		Error: prefix,
	}
}

type errResponse struct {
	Error string
}

func newErrResponse(err error) errResponse {
	return errResponse{Error: err.Error()}
}

func CreatePlotOrder() func(c *gin.Context) {
	return func(c *gin.Context) {
		auth := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

		session, ok := sessions.get(auth)
		if !ok {
			c.IndentedJSON(http.StatusBadRequest, cpoErr("session does not exist", nil))
			return
		}

		cpor := createPlotOrderRequest{}
		if err := c.BindJSON(&cpor); err != nil {
			c.IndentedJSON(http.StatusBadRequest, cpoErr("error marshalling request body", err))
			return
		}

		itv, err := plotor.ParseInterval(cpor.Interval)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, cpoErr("error parsing interval", err))
			return
		}

		plot, err := geometry.FromJSON(cpor.Plot)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, cpoErr("error parsing plot", err))
			return
		}

		po, err := session.PlotOrderer.Create(context.Background(), cpor.Order, plot, itv)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, cpoErr("error creating order", err))
			return
		}

		details, err := po.Order.Details()
		if err != nil {
			res := cpoErr("error retrieving order details", err)
			res.PlotOrderID = po.ID
			c.IndentedJSON(http.StatusBadRequest, res)
			return
		}

		c.IndentedJSON(http.StatusOK, createPlotOrderResponse{
			PlotOrderID: po.ID,
			ClientOrder: details,
		})
	}
}

func GetPlotOrder() func(c *gin.Context) {
	return func(c *gin.Context) {
		auth := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

		session, ok := sessions.get(auth)
		if !ok {
			c.IndentedJSON(http.StatusBadRequest, gpoErr("session does not exist", nil))
			return
		}

		plotOrderID := c.Query("id")
		if plotOrderID == "" {
			c.IndentedJSON(http.StatusBadRequest, gpoErr("id can not be empty", nil))
			return
		}

		po, err := session.PlotOrderer.Get(context.Background(), plotOrderID)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gpoErr("error getting plot order", err))
			return
		}

		details, err := po.Order.Details()
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gpoErr("error retrieving order details", err))
			return
		}

		c.IndentedJSON(http.StatusOK, getPlotOrderResponse{
			PlotOrderID: po.ID,
			Interval:    po.Interval.String(),
			ClientOrder: details,
			Plot: map[string]interface{}{
				"plot json will be here": nil,
			},
			LastTick: po.LastTick,
		})
	}
}

func CancelPlotOrder() func(c *gin.Context) {
	return func(c *gin.Context) {
		auth := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")

		plotOrderID := c.Query("id")
		if plotOrderID == "" {
			c.IndentedJSON(http.StatusBadRequest, newErrResponse(fmt.Errorf("id can not be empty")))
			return
		}

		cancel := false
		cancelParam := c.Query("cancel")
		if cancelParam != "" {
			cancelValue, err := strconv.ParseBool(cancelParam)
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, newErrResponse(fmt.Errorf("invalid cancel param: %v", err)))
				return
			}

			cancel = cancelValue
		}

		session, ok := sessions.get(auth)
		if !ok {
			c.IndentedJSON(http.StatusBadRequest, newErrResponse(fmt.Errorf("session does not exist")))
			return
		}

		err := session.PlotOrderer.Stop(context.Background(), plotOrderID, cancel)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, newErrResponse(fmt.Errorf("error stopping order: %v", err)))
			return
		}

		c.IndentedJSON(http.StatusOK, errResponse{})
	}
}
