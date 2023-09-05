package query

import (
	"strings"

	"github.com/Fabian-G/todotxt/todotxt"
)

func compileStringSearch(query string) Query {
	return func(l *todotxt.List, i *todotxt.Item) bool {
		return strings.Contains(strings.ToLower(i.Description()), strings.ToLower(query))
	}
}
