package qselect

import (
	"fmt"
	"maps"
	"strconv"
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
	wantsContext     bool
}

func (q queryFunc) call(ctx map[string]any, args []any) any {
	if q.wantsContext {
		return q.fn(append([]any{ctx}, args...))
	}
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
		wantsContext:     false,
	},
	"description": {
		fn:               description,
		argTypes:         []DType{QItem},
		resultType:       QString,
		trailingOptional: false,
		injectIt:         true,
		wantsContext:     false,
	},
	"projects": {
		fn:               projects,
		resultType:       QStringSlice,
		argTypes:         []DType{QItem},
		trailingOptional: false,
		injectIt:         true,
		wantsContext:     false,
	},
	"contexts": {
		fn:               contexts,
		resultType:       QStringSlice,
		argTypes:         []DType{QItem},
		trailingOptional: false,
		injectIt:         true,
		wantsContext:     false,
	},
	"dotPrefix": {
		fn:               dotPrefix,
		resultType:       QBool,
		argTypes:         []DType{QString, QString},
		trailingOptional: false,
		injectIt:         false,
		wantsContext:     false,
	},
	"substring": {
		fn:               substring,
		resultType:       QBool,
		argTypes:         []DType{QString, QString},
		trailingOptional: false,
		injectIt:         false,
		wantsContext:     false,
	},
	"date": {
		fn:               date,
		resultType:       QDate,
		argTypes:         []DType{QInt, QInt, QInt},
		trailingOptional: false,
		injectIt:         false,
		wantsContext:     false,
	},
	"dateTag": {
		fn:               dateTag,
		resultType:       QDate,
		argTypes:         []DType{QItem, QString, QDate},
		trailingOptional: true,
		injectIt:         true,
		wantsContext:     false,
	},
	"tag": {
		fn:               tag,
		resultType:       QString,
		argTypes:         []DType{QItem, QString, QString},
		trailingOptional: true,
		injectIt:         true,
		wantsContext:     false,
	},
	"intTag": {
		fn:               intTag,
		resultType:       QInt,
		argTypes:         []DType{QItem, QString, QInt},
		trailingOptional: true,
		injectIt:         true,
		wantsContext:     false,
	},
	"stringListTag": {
		fn:               stringListTag,
		resultType:       QStringSlice,
		argTypes:         []DType{QItem, QString, QStringSlice},
		trailingOptional: true,
		injectIt:         true,
		wantsContext:     false,
	},
}

func RegisterMacro(name, qql string, inTypes []DType, outType DType, injectIt bool) error {
	expectedFreeVars := maps.Clone(defaultFreeVars)
	for i, d := range inTypes {
		expectedFreeVars[fmt.Sprintf("arg%d", i)] = d
	}
	root, err := parseQQLTree(qql, expectedFreeVars, outType)
	if err != nil {
		return err
	}
	qFunc := queryFunc{
		fn: func(args []interface{}) interface{} {
			alpha := args[0].(map[string]any)
			for i, arg := range args[1:] {
				alpha[fmt.Sprintf("arg%d", i)] = arg
			}
			return root.eval(alpha)
		},
		argTypes:         inTypes,
		resultType:       outType,
		trailingOptional: false,
		injectIt:         true,
		wantsContext:     true,
	}
	functions[name] = qFunc
	return nil
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

func dateTag(args []any) any {
	item := args[0].(*todotxt.Item)
	key := args[1].(string)
	defaultDate := time.Time{}
	if len(args) == 3 {
		defaultDate = args[2].(time.Time)
	}

	tags := item.Tags()
	tagValues := tags[key]
	if len(tagValues) == 0 {
		return defaultDate
	}
	dateString := tagValues[0]
	date, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		return defaultDate
	}
	return date
}

func tag(args []any) any {
	item := args[0].(*todotxt.Item)
	key := args[1].(string)
	defaultValue := ""
	if len(args) == 3 {
		defaultValue = args[2].(string)
	}

	tags := item.Tags()
	tagValues := tags[key]
	if len(tagValues) == 0 {
		return defaultValue
	}
	return tagValues[0]
}

func intTag(args []any) any {
	item := args[0].(*todotxt.Item)
	key := args[1].(string)
	defaultInt := 0
	if len(args) == 3 {
		defaultInt = args[2].(int)
	}

	tags := item.Tags()
	tagValues := tags[key]
	if len(tagValues) == 0 {
		return defaultInt
	}
	intString := tagValues[0]
	i, err := strconv.Atoi(intString)
	if err != nil {
		return defaultInt
	}
	return i
}

func stringListTag(args []any) any {
	item := args[0].(*todotxt.Item)
	key := args[1].(string)
	var defaultStringSlice []any = nil
	if len(args) == 3 {
		defaultStringSlice = args[2].([]any)
	}

	tags := item.Tags()
	tagValues := tags[key]
	if len(tagValues) == 0 {
		return defaultStringSlice
	}

	allValues := make([]string, 0, len(tagValues))
	for _, v := range tagValues {
		allValues = append(allValues, strings.Split(v, ",")...)
	}
	return toAnySlice(allValues)
}
