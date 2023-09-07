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
