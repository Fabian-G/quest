package query

import (
	"fmt"
	"strings"
	"time"

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
	argTypes         []DType
	resultType       DType
	trailingOptional bool
	injectIt         bool
}

func (q queryFunc) call(args []any) any {
	return q.fn(args)
}

func (q queryFunc) validate(actual []DType) error {
	for i, t := range q.argTypes {
		switch {
		case len(actual) > len(q.argTypes):
			return fmt.Errorf("expecting parameters %#v, but got %#v", q.argTypes, actual)
		case i >= len(actual) && i == len(q.argTypes)-1 && q.trailingOptional:
			return nil
		case i >= len(actual) && t == QItem && q.injectIt:
			return missingItemError{i}
		case i >= len(actual):
			return fmt.Errorf("expecting parameters %#v, but got %#v", q.argTypes, actual)
		case actual[i] != t && t == QItem && q.injectIt:
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
		argTypes:         []DType{QItem},
		resultType:       QBool,
		trailingOptional: false,
		injectIt:         true,
	},
	"description": {
		fn:               description,
		argTypes:         []DType{QItem},
		resultType:       QString,
		trailingOptional: false,
		injectIt:         true,
	},
	"projects": {
		fn:               projects,
		resultType:       QStringSlice,
		argTypes:         []DType{QItem},
		trailingOptional: false,
		injectIt:         true,
	},
	"contexts": {
		fn:               contexts,
		resultType:       QStringSlice,
		argTypes:         []DType{QItem},
		trailingOptional: false,
		injectIt:         true,
	},
	"dotPrefix": {
		fn:               dotPrefix,
		resultType:       QBool,
		argTypes:         []DType{QString, QString},
		trailingOptional: false,
		injectIt:         false,
	},
	"substring": {
		fn:               substring,
		resultType:       QBool,
		argTypes:         []DType{QString, QString},
		trailingOptional: false,
		injectIt:         false,
	},
	"date": {
		fn:               date,
		resultType:       QDate,
		argTypes:         []DType{QInt, QInt, QInt},
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

func date(args []any) any {
	year := args[0].(int)
	month := args[1].(int)
	day := args[2].(int)

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
