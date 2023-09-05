package query

import (
	"fmt"

	"github.com/Fabian-G/todotxt/todotxt"
)

type Type int

const (
	Guess Type = iota
	FOL
	Range
	StringSearch
)

type Query func(*todotxt.List, *todotxt.Item) bool

func (q Query) Filter(l *todotxt.List) []*todotxt.Item {
	allTasks := l.Tasks()
	matches := make([]*todotxt.Item, 0)
	for _, t := range allTasks {
		if q(l, t) {
			matches = append(matches, t)
		}
	}
	return matches
}

func Compile(query string, typ Type) (Query, error) {
	switch typ {
	case FOL:
		return compileFOL(query)
	case Range:
		return compileRange(query)
	case StringSearch:
		return compileStringSearch(query), nil
	case Guess:
		q, err := compileFOL(query)
		if err == nil {
			return q, nil
		}
		q, err = compileRange(query)
		if err == nil {
			return q, nil
		}
		return compileStringSearch(query), nil
	}
	return nil, fmt.Errorf("unknown query type")
}
