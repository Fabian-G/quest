package qselect

import (
	"fmt"
	"maps"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Fabian-G/quest/todotxt"
)

var maxTime = time.Unix(1<<63-62135596801, 999999999)

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
	"line": {
		fn:               line,
		argTypes:         []DType{QItem},
		resultType:       QInt,
		trailingOptional: false,
		injectIt:         true,
		wantsContext:     true,
	},
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
	"creation": {
		fn:               creation,
		argTypes:         []DType{QItem, QDate},
		resultType:       QDate,
		trailingOptional: true,
		injectIt:         true,
		wantsContext:     false,
	},
	"completion": {
		fn:               completion,
		argTypes:         []DType{QItem, QDate},
		resultType:       QDate,
		trailingOptional: true,
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
	"priority": {
		fn:               priority,
		resultType:       QPriority,
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
	"ymd": {
		fn:               ymd,
		resultType:       QDate,
		argTypes:         []DType{QInt, QInt, QInt},
		trailingOptional: false,
		injectIt:         false,
		wantsContext:     false,
	},
	"date": {
		fn:               toDate,
		resultType:       QDate,
		argTypes:         []DType{QString, QDate},
		trailingOptional: true,
		injectIt:         false,
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
	"list": {
		fn:               toList,
		resultType:       QStringSlice,
		argTypes:         []DType{QString},
		trailingOptional: false,
		injectIt:         true,
		wantsContext:     false,
	},
	"shell": {
		fn:               shell,
		resultType:       QString,
		argTypes:         []DType{QItem, QString},
		trailingOptional: false,
		injectIt:         true,
		wantsContext:     true,
	},
	"command": {
		fn:               command,
		resultType:       QString,
		argTypes:         []DType{QItem, QString},
		trailingOptional: false,
		injectIt:         true,
		wantsContext:     true,
	},
	"int": {
		fn:               toInt,
		resultType:       QInt,
		argTypes:         []DType{QString, QInt},
		trailingOptional: true,
		injectIt:         false,
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
				argName := fmt.Sprintf("arg%d", i)
				before := alpha[argName]
				defer func() {
					alpha[argName] = before
				}()
				alpha[argName] = arg
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

func line(args []any) any {
	list := args[0].(map[string]any)["_list"].(*todotxt.List)
	item := args[1].(*todotxt.Item)
	return list.LineOf(item)
}

func done(args []any) any {
	item := args[0].(*todotxt.Item)
	return item.Done()
}

func description(args []any) any {
	item := args[0].(*todotxt.Item)
	return item.Description()
}

func creation(args []any) any {
	item := args[0].(*todotxt.Item)
	defaultCreation := time.Time{}
	if len(args) == 2 {
		defaultCreation = args[1].(time.Time)
	}
	creation := item.CreationDate()
	if creation == nil {
		return defaultCreation
	}
	return *creation
}

func completion(args []any) any {
	item := args[0].(*todotxt.Item)
	defaultCompletion := maxTime
	if len(args) == 2 {
		defaultCompletion = args[1].(time.Time)
	}
	completion := item.CompletionDate()
	if completion == nil {
		return defaultCompletion
	}
	return *completion
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

func priority(args []any) any {
	item := args[0].(*todotxt.Item)
	return item.Priority()
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

func ymd(args []any) any {
	year := args[0].(int)
	month := args[1].(int)
	day := args[2].(int)

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

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

func toList(args []any) any {
	list := args[0].(string)
	if len(strings.TrimSpace(list)) == 0 {
		return []any{}
	}
	return toAnySlice(strings.Split(list, ","))
}

func toInt(args []any) any {
	intString := args[0].(string)
	defaultInt := 0
	if len(args) == 2 {
		defaultInt = args[1].(int)
	}
	result, err := strconv.Atoi(intString)
	if err != nil {
		return defaultInt
	}
	return result
}

func toDate(args []any) any {
	dateString := args[0].(string)
	defaultDate := time.Time{}
	if len(args) == 2 {
		defaultDate = args[1].(time.Time)
	}
	result, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		return defaultDate
	}
	return result
}

func shell(args []any) any {
	alpha := args[0].(map[string]any)
	item := args[1].(*todotxt.Item)
	command := args[2].(string)
	cmd := exec.Command("bash", "-c", command)
	buffer := strings.Builder{}
	if err := todotxt.DefaultJsonEncoder.Encode(&buffer, alpha["_list"].(*todotxt.List), []*todotxt.Item{item}); err != nil {
		panic(err) // should never happen, because we write to a buffer
	}
	itemJson := strings.Trim(strings.TrimSpace(buffer.String()), "[]")
	cmd.Stdin = strings.NewReader(itemJson)
	outBuffer := strings.Builder{}
	errBuffer := strings.Builder{}
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	err := cmd.Run()
	if err != nil {
		panic(fmt.Errorf("shell command returned an error: %s\nErrOut: %s", err, errBuffer.String()))
	}
	return strings.TrimSpace(outBuffer.String())
}

func command(args []any) any {
	alpha := args[0].(map[string]any)
	item := args[1].(*todotxt.Item)
	command := args[2].(string)
	fullCommand, err := exec.LookPath(command)
	if err != nil {
		panic(fmt.Errorf("could not run command function in query, because executable was not found: %w", err))
	}
	cmd := exec.Command(fullCommand)
	buffer := strings.Builder{}
	if err = todotxt.DefaultJsonEncoder.Encode(&buffer, alpha["_list"].(*todotxt.List), []*todotxt.Item{item}); err != nil {
		panic(err) // should never happen, because we write to a buffer
	}
	itemJson := strings.Trim(strings.TrimSpace(buffer.String()), "[]")
	cmd.Stdin = strings.NewReader(itemJson)
	outBuffer := strings.Builder{}
	errBUffer := strings.Builder{}
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBUffer

	err = cmd.Run()
	if err != nil {
		panic(fmt.Errorf("command returned an error: %w\nErrOut: %s", err, errBUffer.String()))
	}
	return strings.TrimSpace(outBuffer.String())
}
