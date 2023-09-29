package qselect

import (
	"maps"
	"math"
	"time"

	"github.com/Fabian-G/quest/todotxt"
)

type Func func(*todotxt.List, *todotxt.Item) bool

func (q Func) Filter(l *todotxt.List) []*todotxt.Item {
	allTasks := l.Tasks()
	matches := make([]*todotxt.Item, 0)
	for _, t := range allTasks {
		if q(l, t) {
			matches = append(matches, t)
		}
	}
	return matches
}

func And(fns ...Func) Func {
	return func(l *todotxt.List, i *todotxt.Item) bool {
		for _, fn := range fns {
			if !fn(l, i) {
				return false
			}
		}
		return true
	}
}

var defaultFreeVars = idSet{
	"it":      QItem,
	"items":   QItemSlice,
	"today":   QDate,
	"maxInt":  QInt,
	"minInt":  QInt,
	"maxDate": QDate,
	"minDate": QDate,
}

func buildFreeVars(universe *todotxt.List, item *todotxt.Item) map[string]any {
	alpha := make(map[string]any)
	alpha["it"] = item
	alpha["items"] = toAnySlice(universe.Tasks())
	now := time.Now()
	alpha["today"] = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	alpha["maxInt"] = math.MaxInt
	alpha["minInt"] = math.MinInt
	alpha["minDate"] = time.Time{}
	alpha["maxDate"] = maxTime
	alpha["_list"] = universe
	return alpha
}

func CompileQuery(query string) (Func, error) {
	q, err := CompileQQL(query)
	if err == nil {
		return q, nil
	}
	q, err = compileRange(query)
	if err == nil {
		return q, nil
	}
	return compileStringSearch(query), nil
}

func CompileQQL(query string) (Func, error) {
	root, err := parseQQLTree(query, maps.Clone(defaultFreeVars), QBool)
	if err != nil {
		return nil, err
	}
	evalFunc := func(universe *todotxt.List, it *todotxt.Item) bool {
		return root.eval(buildFreeVars(universe, it)).(bool)
	}
	return evalFunc, nil
}

func CompileRange(query string) (Func, error) {
	return compileRange(query)
}

func CompileWordSearch(query string) (Func, error) {
	return compileStringSearch(query), nil
}
