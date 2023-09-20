package hook

import (
	"strconv"
	"strings"
	"time"
)

type durationUnit int

const (
	day durationUnit = iota
	week
	month
	year
)

var unitMap = map[string]durationUnit{
	"d":      day,
	"days":   day,
	"w":      week,
	"weeks":  week,
	"m":      month,
	"months": month,
	"y":      year,
	"years":  year,
}

type duration struct {
	span int
	unit durationUnit
}

func (d duration) addTo(t time.Time) time.Time {
	switch d.unit {
	case day:
		return time.Date(t.Year(), t.Month(), t.Day()+d.span, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	case week:
		return time.Date(t.Year(), t.Month(), t.Day()+d.span*7, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	case month:
		return time.Date(t.Year(), t.Month()+time.Month(d.span), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	case year:
		return time.Date(t.Year()+d.span, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	default:
		panic("unknown unit")
	}
}

func (d duration) abs() duration {
	if d.span < 0 {
		return duration{-d.span, d.unit}
	}
	return duration{d.span, d.unit}
}

func parseDuration(durationS string) (duration, error) {
	var unit durationUnit
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
		return duration{}, err
	}
	return duration{span, unit}, nil
}
