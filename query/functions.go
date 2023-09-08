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

func (q queryFunc) validate(actual []dType) error {
	for i, t := range q.argTypes {
		switch {
		case len(actual) > len(q.argTypes):
			return fmt.Errorf("expecting parameters %#v, but got %#v", q.argTypes, actual)
		case i >= len(actual) && i == len(q.argTypes)-1 && q.trailingOptional:
			return nil
		case i >= len(actual) && t == qItem && q.injectIt:
			return missingItemError{i}
		case i >= len(actual):
			return fmt.Errorf("expecting parameters %#v, but got %#v", q.argTypes, actual)
		case actual[i] != t && t == qItem && q.injectIt:
			return missingItemError{i}
		case actual[i] != t:
			return fmt.Errorf("expecting parameters %#v, but got %#v", q.argTypes, actual)
		}
	}
	return nil
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
	"itemEq": {
		fn:               itemEq,
		resultType:       qBool,
		argTypes:         []dType{qItem, qItem},
		trailingOptional: false,
		injectIt:         true,
	},
}

func done(args []any) any {
	item := args[0].(*todotxt.Item)
	return item.Done()
}

func projects(args []any) any {
	item := args[0].(*todotxt.Item)
	projects := item.Projects()
	projStrings := make([]any, 0, len(projects))
	for _, p := range projects {
		projStrings = append(projStrings, p.String())
	}
	return projStrings
}

func contains(args []any) any {
	slice := make([]string, 0, len(args[0].([]any)))
	for _, e := range args[0].([]any) {
		slice = append(slice, e.(string))
	}
	elem := args[1].(string)
	return slices.Contains(slice, elem)
}

func itemEq(args []any) any {
	i1 := args[0].(*todotxt.Item)
	i2 := args[1].(*todotxt.Item)
	return i1 == i2
}
