package query

import (
	"fmt"
	"slices"

	"github.com/Fabian-G/quest/todotxt"
)

type missingItemError struct {
	position int
}

func (m missingItemError) Error() string {
	return fmt.Sprintf("item missing in argslist at position %d", m.position)
}

type queryFunc struct {
	fn               func([]any) any
	argTypes         []dType
	resultType       dType
	trailingOptional bool
	injectIt         bool
}

func (q queryFunc) call(args []any) any {
	return q.fn(args)
}

func (q queryFunc) validate(actual []dType) (dType, error) {
	offset := 0
	injectItPosition := -1
	for i, t := range q.argTypes {
		i := i - offset
		switch {
		case i >= len(actual) && !q.trailingOptional && q.injectIt && t == qItem && injectItPosition == -1:
			offset = 1
			injectItPosition = i
			continue
		case i >= len(actual):
			if i == len(q.argTypes)-1 && q.trailingOptional {
				return q.resultType, nil
			}
			return qError, fmt.Errorf("expecting parameters %#v, but got %#v", q.argTypes, actual)
		case actual[i] == t:
			continue
		case actual[i] != t && q.injectIt && t == qItem && injectItPosition == -1:
			offset = 1
			injectItPosition = i
			continue
		case actual[i] != t:
			return qError, fmt.Errorf("expecting parameters %#v, but got %#v", q.argTypes, actual)
		}
	}
	if injectItPosition != -1 {
		return qError, missingItemError{injectItPosition}
	}
	return q.resultType, nil
}

var functions = map[string]queryFunc{
	"done": {
		fn:               done,
		argTypes:         []dType{qItem},
		resultType:       qBool,
		trailingOptional: false,
		injectIt:         true,
	},
	"projects": {
		fn:               projects,
		resultType:       qStringSlice,
		argTypes:         []dType{qItem},
		trailingOptional: false,
		injectIt:         true,
	},
	"contains": {
		fn:               contains,
		resultType:       qBool,
		argTypes:         []dType{qStringSlice, qString},
		trailingOptional: false,
		injectIt:         false,
	},
}

func done(args []any) any {
	item := args[0].(*todotxt.Item)
	return item.Done()
}

func projects(args []any) any {
	item := args[0].(*todotxt.Item)
	projects := item.Projects()
	projStrings := make([]string, 0, len(projects))
	for _, p := range projects {
		projStrings = append(projStrings, p.String())
	}
	return projStrings
}

func contains(args []any) any {
	slice := args[0].([]string)
	elem := args[1].(string)
	return slices.Contains(slice, elem)
}
