package query

import (
	"errors"
	"fmt"
	"slices"

	"github.com/Fabian-G/todotxt/todotxt"
)

type queryFunc func([]any) any

var functions = map[string]queryFunc{
	"done":     done,
	"projects": projects,
	"contains": contains,
}

func typecheck(funcName string, argsTypes []string) (string, error) {
	switch funcName {
	case "done":
		return "bool", assertTypes([]string{"item"}, argsTypes)
	case "projects":
		return "[]string", assertTypes([]string{"item"}, argsTypes)
	case "contains":
		return "bool", assertTypes([]string{"[]string, string"}, argsTypes)
	}
	return "", errors.New("Invalid function name")
}

func assertTypes(expectedTypes []string, argTypes []string) error {
	if len(expectedTypes) != len(argTypes) {
		return fmt.Errorf("expecting parameters %#v, but got %#v", expectedTypes, argTypes)
	}
	for i, typ := range expectedTypes {
		if argTypes[i] != typ {
			return fmt.Errorf("expecting parameters %#v, but got %#v", expectedTypes, argTypes)
		}
	}
	return nil
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
