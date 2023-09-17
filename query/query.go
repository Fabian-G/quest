package query

import (
	"maps"
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

var DefaultCompiler = Compiler{}
var defaultFreeVars = idSet{
	"it": QItem, "items": QItemSlice, "today": QDate,
}

type Compiler struct {
	TagTypes map[string]DType
}

func (c *Compiler) CompileQuery(query string) (Func, error) {
	q, err := c.CompileQQL(query)
	if err == nil {
		return q, nil
	}
	q, err = compileRange(query)
	if err == nil {
		return q, nil
	}
	return compileStringSearch(query), nil
}

func (c *Compiler) CompileQQL(query string) (Func, error) {
	root, err := parseQQLTree(query, maps.Clone(defaultFreeVars), QBool)
	if err != nil {
		return nil, err
	}
	evalFunc := func(universe *todotxt.List, it *todotxt.Item) bool {
		alpha := make(map[string]any)
		alpha["it"] = it
		alpha["items"] = toAnySlice(universe.Tasks())
		now := time.Now()
		alpha["today"] = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return root.eval(alpha).(bool)
	}
	return evalFunc, nil
}

func (c *Compiler) CompileRange(query string) (Func, error) {
	return compileRange(query)
}

func (c *Compiler) CompileWordSearch(query string) (Func, error) {
	return compileStringSearch(query), nil
}

func (c *Compiler) CompileSortFunc(sort string) (func(*todotxt.Item, *todotxt.Item) int, error) {
	return sortFunc(sort, c.TagTypes)
}
