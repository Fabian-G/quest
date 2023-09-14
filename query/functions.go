package query

import (
	"fmt"
	"strings"

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
	"description": {
		fn:               description,
		argTypes:         []dType{qItem},
		resultType:       qString,
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
	"contexts": {
		fn:               contexts,
		resultType:       qStringSlice,
		argTypes:         []dType{qItem},
		trailingOptional: false,
		injectIt:         true,
	},
	"dotPrefix": {
		fn:               dotPrefix,
		resultType:       qBool,
		argTypes:         []dType{qString, qString},
		trailingOptional: false,
		injectIt:         false,
	},
	"substring": {
		fn:               substring,
		resultType:       qBool,
		argTypes:         []dType{qString, qString},
		trailingOptional: false,
		injectIt:         false,
	},
}

func done(args []any) any {
	item := args[0].(*todotxt.Item)
	return item.Done()
}

func description(args []any) any {
	item := args[0].(*todotxt.Item)
	return item.Description()
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

func contexts(args []any) any {
	item := args[0].(*todotxt.Item)
	contexts := item.Contexts()
	contextStrings := make([]any, 0, len(contexts))
	for _, p := range contexts {
		contextStrings = append(contextStrings, p.String())
	}
	return contextStrings
}

func dotPrefix(args []any) any {
	s1 := args[0].(string)
	s2 := args[1].(string)

	if !strings.HasPrefix(s1, s2) {
		return false
	}
	if len(s1) == len(s2) {
		return true
	}
	return s1[len(s2)] == '.'
}

func substring(args []any) any {
	s1 := args[0].(string)
	s2 := args[1].(string)

	return strings.Contains(s1, s2)
}
