package query

import (
	"fmt"
	"slices"

	"github.com/Fabian-G/quest/todotxt"
)

type queryFunc func([]any) any

var functions = map[string]queryFunc{
	"done":     done,
	"projects": projects,
	"contains": contains,
}

func funcType(funcName string, argsTypes []dType) (dType, error) {
	switch funcName {
	case "done":
		return qBool, assertTypes([]dType{qItem}, argsTypes)
	case "projects":
		return qStringSlice, assertTypes([]dType{qItem}, argsTypes)
	case "contains":
		return qBool, assertTypes([]dType{qStringSlice, qString}, argsTypes)
	}
	return qError, fmt.Errorf("unknown function: %s", funcName)
}

func assertTypes(expectedTypes []dType, argTypes []dType) error {
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
