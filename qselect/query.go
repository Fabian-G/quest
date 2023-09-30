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
	"it":       QItem,
	"items":    QItemSlice,
	"today":    QDate,
	"maxInt":   QInt,
	"minInt":   QInt,
	"maxDate":  QDate,
	"minDate":  QDate,
	"prioA":    QPriority,
	"prioB":    QPriority,
	"prioC":    QPriority,
	"prioD":    QPriority,
	"prioE":    QPriority,
	"prioF":    QPriority,
	"prioG":    QPriority,
	"prioH":    QPriority,
	"prioI":    QPriority,
	"prioJ":    QPriority,
	"prioK":    QPriority,
	"prioL":    QPriority,
	"prioM":    QPriority,
	"prioN":    QPriority,
	"prioO":    QPriority,
	"prioP":    QPriority,
	"prioQ":    QPriority,
	"prioR":    QPriority,
	"prioS":    QPriority,
	"prioT":    QPriority,
	"prioU":    QPriority,
	"prioV":    QPriority,
	"prioW":    QPriority,
	"prioX":    QPriority,
	"prioY":    QPriority,
	"prioZ":    QPriority,
	"prioNone": QPriority,
}

var constants = map[string]any{
	"maxInt":   math.MaxInt,
	"minInt":   math.MinInt,
	"maxDate":  maxTime,
	"minDate":  time.Time{},
	"prioA":    todotxt.PrioA,
	"prioB":    todotxt.PrioB,
	"prioC":    todotxt.PrioC,
	"prioD":    todotxt.PrioD,
	"prioE":    todotxt.PrioE,
	"prioF":    todotxt.PrioF,
	"prioG":    todotxt.PrioG,
	"prioH":    todotxt.PrioH,
	"prioI":    todotxt.PrioI,
	"prioJ":    todotxt.PrioJ,
	"prioK":    todotxt.PrioK,
	"prioL":    todotxt.PrioL,
	"prioM":    todotxt.PrioM,
	"prioN":    todotxt.PrioN,
	"prioO":    todotxt.PrioO,
	"prioP":    todotxt.PrioP,
	"prioQ":    todotxt.PrioQ,
	"prioR":    todotxt.PrioR,
	"prioS":    todotxt.PrioS,
	"prioT":    todotxt.PrioT,
	"prioU":    todotxt.PrioU,
	"prioV":    todotxt.PrioV,
	"prioW":    todotxt.PrioW,
	"prioX":    todotxt.PrioX,
	"prioY":    todotxt.PrioY,
	"prioZ":    todotxt.PrioZ,
	"prioNone": todotxt.PrioNone,
}

func buildFreeVars(universe *todotxt.List, item *todotxt.Item) map[string]any {
	alpha := maps.Clone(constants)
	alpha["it"] = item
	alpha["items"] = toAnySlice(universe.Tasks())
	now := time.Now()
	alpha["today"] = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
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
