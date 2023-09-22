package qduration

import (
	"strconv"
	"strings"
	"time"
)

type DurationUnit int

const (
	Day DurationUnit = iota
	Week
	Month
	Year
)

var unitMap = map[string]DurationUnit{
	"d":      Day,
	"days":   Day,
	"w":      Week,
	"weeks":  Week,
	"m":      Month,
	"months": Month,
	"y":      Year,
	"years":  Year,
}

type Duration struct {
	span int
	unit DurationUnit
}

func (d Duration) AddTo(t time.Time) time.Time {
	switch d.unit {
	case Day:
		return time.Date(t.Year(), t.Month(), t.Day()+d.span, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	case Week:
		return time.Date(t.Year(), t.Month(), t.Day()+d.span*7, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	case Month:
		return time.Date(t.Year(), t.Month()+time.Month(d.span), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	case Year:
		return time.Date(t.Year()+d.span, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	default:
		panic("unknown unit")
	}
}

func (d Duration) SubFrom(t time.Time) time.Time {
	switch d.unit {
	case Day:
		return time.Date(t.Year(), t.Month(), t.Day()-d.span, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	case Week:
		return time.Date(t.Year(), t.Month(), t.Day()-d.span*7, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	case Month:
		return time.Date(t.Year(), t.Month()-time.Month(d.span), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	case Year:
		return time.Date(t.Year()-d.span, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	default:
		panic("unknown unit")
	}
}

func (d Duration) Abs() Duration {
	if d.span < 0 {
		return Duration{-d.span, d.unit}
	}
	return Duration{d.span, d.unit}
}

func (d Duration) Days() int {
	switch d.unit {
	case Day:
		return d.span
	case Week:
		return d.span * 7
	case Month:
		return d.span * 30
	case Year:
		return d.span * 365
	}
	return d.span
}

func Parse(durationS string) (Duration, error) {
	var unit DurationUnit
	for u := range unitMap {
		if strings.HasSuffix(durationS, u) {
			unit = unitMap[u]
			durationS = strings.TrimSuffix(durationS, u)
		}
	}
	var span int
	if len(durationS) == 0 {
		span = 0
	}
	span, err := strconv.Atoi(durationS)
	if err != nil {
		return Duration{}, err
	}
	return Duration{span, unit}, nil
}
