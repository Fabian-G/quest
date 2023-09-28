// stolen from https://github.com/dustin/go-humanize
package qprojection

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// Seconds-based time units
const (
	day      = 24 * time.Hour
	week     = 7 * day
	month    = 30 * day
	year     = 12 * month
	longTime = 37 * year
)

// humanTime formats a time into a relative string.
//
// humanTime(jsomeT) -> "3 weeks ago"
func humanTime(then time.Time) string {
	then = time.Date(then.Year(), then.Month(), then.Day(), 0, 0, 0, 0, time.UTC)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return relTime(then, today, "ago", "in")
}

// A relTimeMagnitude struct contains a relative time point at which
// the relative format of time will switch to a new format string.  A
// slice of these in ascending order by their "D" field is passed to
// CustomRelTime to format durations.
//
// The Format field is a string that may contain a "%s" which will be
// replaced with the appropriate signed label (e.g. "ago" or "from
// now") and a "%d" that will be replaced by the quantity.
//
// The DivBy field is the amount of time the time difference must be
// divided by in order to display correctly.
//
// e.g. if D is 2*time.Minute and you want to display "%d minutes %s"
// DivBy should be time.Minute so whatever the duration is will be
// expressed in minutes.
type relTimeMagnitude struct {
	D       time.Duration
	AFormat string
	BFormat string
	DivBy   time.Duration
}

var defaultMagnitudes = []relTimeMagnitude{
	{day, "today", "today", 1},
	{2 * day, "1 day %s", "%s 1 day", 1},
	{month, "%d days %s", "%s %d days", day},
	{2 * month, "1 month %s", "%s 1 month", 1},
	{year, "%d months %s", "%s %d months", month},
	{18 * month, "1 year %s", "%s 1 year", 1},
	{2 * year, "2 years %s", "%s %d years", 1},
	{math.MaxInt64, "a long while %s", "%s a long while", 1},
}

// relTime formats a time into a relative string.
//
// It takes two times and two labels.  In addition to the generic time
// delta string (e.g. 5 minutes), the labels are used applied so that
// the label corresponding to the smaller time is applied.
//
// relTime(timeInPast, timeInFuture, "earlier", "later") -> "3 weeks earlier"
func relTime(a, b time.Time, albl, blbl string) string {
	return customRelTime(a, b, albl, blbl, defaultMagnitudes)
}

// customRelTime formats a time into a relative string.
//
// It takes two times two labels and a table of relative time formats.
// In addition to the generic time delta string (e.g. 5 minutes), the
// labels are used applied so that the label corresponding to the
// smaller time is applied.
func customRelTime(a, b time.Time, albl, blbl string, magnitudes []relTimeMagnitude) string {
	lbl := albl
	diff := b.Sub(a)

	if a.After(b) {
		lbl = blbl
		diff = a.Sub(b)
	}

	n := sort.Search(len(magnitudes), func(i int) bool {
		return magnitudes[i].D > diff
	})

	if n >= len(magnitudes) {
		n = len(magnitudes) - 1
	}
	mag := magnitudes[n]
	args := []interface{}{}
	escaped := false
	format := mag.AFormat
	if a.After(b) {
		format = mag.BFormat
	}
	for _, ch := range format {
		if escaped {
			switch ch {
			case 's':
				args = append(args, lbl)
			case 'd':
				args = append(args, diff/mag.DivBy)
			}
			escaped = false
		} else {
			escaped = ch == '%'
		}
	}
	return fmt.Sprintf(format, args...)
}
