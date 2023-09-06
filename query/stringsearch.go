package query

import (
	"strings"

	"github.com/Fabian-G/quest/todotxt"
)

func compileStringSearch(query string) Func {
	return func(l *todotxt.List, i *todotxt.Item) bool {
		return strings.Contains(strings.ToLower(i.Description()), strings.ToLower(query))
	}
}
