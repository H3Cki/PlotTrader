package geometry

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	KEY_LINE              = "line"
	KEY_LOG_LINE          = "log_line"
	KEY_ABSOLUTE_OFFSET   = "absolute_offset"
	KEY_PERCENTAGE_OFFSET = "percentage_offset"
	KEY_MIN               = "min"
	KEY_MAX               = "max"
	KEY_SCHEDULE          = "schedule"
)

// plotJSON is a general structure holding plot type and unparse arguments for that type
type plotJSON struct {
	Type string
	Args json.RawMessage
}

// linePlotJSON is a structure holding arguments for Line and LogLine
type linePlotJSON struct {
	P0, P1                  Point
	ExtendLeft, ExtendRight bool
}

// oggsetPlotJSON is a structure holding arguments for AbsoluteOffset and PercentageOffset
type offsetPlotJSON struct {
	Value float64
	Plot  plotJSON
}

// oggsetPlotJSON is a structure holding arguments for Schedule
type schedulePlotJSON struct {
	Since, Until time.Time
	Plot         plotJSON
}

// oggsetPlotJSON is a structure holding arguments for Min and Max
type minMaxPlotJSON struct {
	Plots []plotJSON
}

func FromJSON(data []byte) (Plot, error) {
	pj := plotJSON{}
	if err := json.Unmarshal(data, &pj); err != nil {
		return nil, fmt.Errorf("error unmarshalling json: %v", err)
	}

	return parsePlot(pj)
}

func parsePlot(pj plotJSON) (Plot, error) {
	args := pj.Args

	switch pj.Type {
	case KEY_LINE:
		lineJSON := linePlotJSON{}
		if err := json.Unmarshal(args, &lineJSON); err != nil {
			return nil, err
		}

		return NewLine(lineJSON.P0, lineJSON.P1, lineJSON.ExtendLeft, lineJSON.ExtendRight)
	case KEY_LOG_LINE:
		lineJSON := linePlotJSON{}
		if err := json.Unmarshal(args, &lineJSON); err != nil {
			return nil, err
		}

		return NewLogLine(lineJSON.P0, lineJSON.P1, lineJSON.ExtendLeft, lineJSON.ExtendRight)
	case KEY_ABSOLUTE_OFFSET:
		offsetJSON := offsetPlotJSON{}
		if err := json.Unmarshal(args, &offsetJSON); err != nil {
			return nil, err
		}

		plotToOffset, err := parsePlot(offsetJSON.Plot)
		if err != nil {
			return nil, err
		}

		return NewOffsetPlot(plotToOffset, NewAbsoluteOffset(offsetJSON.Value)), nil
	case KEY_PERCENTAGE_OFFSET:
		offsetJSON := offsetPlotJSON{}
		if err := json.Unmarshal(args, &offsetJSON); err != nil {
			return nil, err
		}

		plotToOffset, err := parsePlot(offsetJSON.Plot)
		if err != nil {
			return nil, err
		}

		return NewOffsetPlot(plotToOffset, NewPercentageOffset(offsetJSON.Value)), nil
	case KEY_SCHEDULE:
		scheduleJSON := schedulePlotJSON{}
		if err := json.Unmarshal(args, &scheduleJSON); err != nil {
			return nil, err
		}

		plotToSchedule, err := parsePlot(scheduleJSON.Plot)
		if err != nil {
			return nil, err
		}

		return NewSchedule(scheduleJSON.Since, scheduleJSON.Until, plotToSchedule), nil
	case KEY_MIN:
		minmaxJSON := minMaxPlotJSON{}
		if err := json.Unmarshal(args, &minmaxJSON); err != nil {
			return nil, err
		}

		plotsToMin := []Plot{}
		for _, pjMin := range minmaxJSON.Plots {
			plotToMin, err := parsePlot(pjMin)
			if err != nil {
				return nil, err
			}

			plotsToMin = append(plotsToMin, plotToMin)
		}

		return NewMin(plotsToMin)
	case KEY_MAX:
		minmaxJSON := minMaxPlotJSON{}
		if err := json.Unmarshal(args, &minmaxJSON); err != nil {
			return nil, err
		}

		plotsToMin := []Plot{}
		for _, pjMin := range minmaxJSON.Plots {
			plotToMin, err := parsePlot(pjMin)
			if err != nil {
				return nil, err
			}

			plotsToMin = append(plotsToMin, plotToMin)
		}

		return NewMax(plotsToMin)
	}

	return nil, fmt.Errorf("unknown plot name %s", pj.Type)
}
