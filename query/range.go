package query

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/Fabian-G/quest/todotxt"
)

func compileRange(query string) (Func, error) {
	ranges := strings.Split(query, ",")
	if len(ranges) == 0 {
		return nil, errors.New("empty range is not a valid query")
	}
	type interval struct {
		begin int
		end   int
	}
	intervals := make([]interval, 0, len(ranges))
	for _, r := range ranges {
		bounds := strings.Split(r, "-")
		switch {
		case len(bounds) == 0:
			continue
		case len(bounds) > 2:
			return nil, errors.New("only one minus sign per interval allowed")
		case len(bounds) == 1:
			idx, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("the range bounds must be valid integers: %w", err)
			}
			intervals = append(intervals, interval{idx, idx})
		case len(bounds) == 2:
			var leftBound int
			var err error
			if len(strings.TrimSpace(bounds[0])) == 0 {
				leftBound = 0
			} else {
				leftBound, err = strconv.Atoi(strings.TrimSpace(bounds[0]))
				if err != nil {
					return nil, fmt.Errorf("the range bounds must be valid integers: %w", err)
				}
			}
			var rightBound int
			if len(strings.TrimSpace(bounds[1])) == 0 {
				rightBound = math.MaxInt
			} else {
				rightBound, err = strconv.Atoi(strings.TrimSpace(bounds[1]))
				if err != nil {
					return nil, fmt.Errorf("the range bounds must be valid integers: %w", err)
				}
			}
			intervals = append(intervals, interval{leftBound, rightBound})
		}
	}

	return func(l *todotxt.List, i *todotxt.Item) bool {
		idx := l.IndexOf(i)
		return slices.ContainsFunc(intervals, func(i interval) bool {
			return i.begin <= idx && idx <= i.end
		})
	}, nil

}
