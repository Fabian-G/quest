package qselect

import (
	"errors"
	"slices"
	"testing"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

func TestQueryFunc_validate(t *testing.T) {
	testcases := map[string]struct {
		fn      queryFunc
		actual  []DType
		allowed bool
	}{
		"one missing it": {
			fn: queryFunc{
				argTypes:         []DType{QBool, QItem, QBool},
				trailingOptional: false,
				injectIt:         true,
			},
			actual:  []DType{QBool, QBool},
			allowed: true,
		},
		"Exact args": {
			fn: queryFunc{
				argTypes:         []DType{QBool, QItem, QBool},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []DType{QBool, QItem, QBool},
			allowed: true,
		},
		"Inserting it leads to too many args": {
			fn: queryFunc{
				argTypes:         []DType{QItem, QItem, QItem},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []DType{QItem, QItem, QBool},
			allowed: false,
		},
		"It on zero args can be inserted": {
			fn: queryFunc{
				argTypes:         []DType{QItem},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []DType{},
			allowed: true,
		},
		"Trailing Optional can be omitted": {
			fn: queryFunc{
				argTypes:         []DType{QBool},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []DType{},
			allowed: true,
		},
		"Trailing Optional can be omitted (with many args)": {
			fn: queryFunc{
				argTypes:         []DType{QBool, QBool, QBool},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []DType{QBool, QBool},
			allowed: true,
		},
		"Only one trailing arg can be omitted": {
			fn: queryFunc{
				argTypes:         []DType{QBool, QBool, QBool},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []DType{QBool},
			allowed: false,
		},
		"Injecting it and trailing optional can happen at the same time": {
			fn: queryFunc{
				argTypes:         []DType{QBool, QItem, QBool, QBool},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []DType{QBool, QBool},
			allowed: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			err := tc.fn.validate(tc.actual)
			var missingItemError missingItemError
			if errors.As(err, &missingItemError) {
				tc.actual = slices.Insert(tc.actual, missingItemError.position, QItem)
			}
			err = tc.fn.validate(tc.actual)
			if tc.allowed {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func Test_CanRegisterMacros(t *testing.T) {
	err := RegisterMacro("blocked",
		`exists i in items: (exists pre in stringListTag(arg0, "after"): tag(i, "id") == pre) && !done(i)`,
		[]DType{QItem},
		QBool,
		true,
	)
	assert.Nil(t, err)
	defer func() {
		delete(functions, "blocked")
	}()

	list := listFromString(t, `
	a precondition id:1
	a blocked item after:1,2
	x a completed precondition id:2
	an item after completed after:2
	`)

	query, err := CompileQQL("!done && !blocked")
	assert.Nil(t, err)

	matches := query.Filter(list)
	assert.Len(t, matches, 2)
	assert.Equal(t, list.Get(1).Description(), matches[0].Description())
	assert.Equal(t, list.Get(4).Description(), matches[1].Description())
}

func Test_AlphaIsRestoredAfterMacroExecution(t *testing.T) {
	RegisterMacro("testMacro", `true`, []DType{QItem}, QBool, true)
	defer func() {
		delete(functions, "testMacro")
	}()
	// if testMacro leaks arg0 into the outer scope this query will always be false if "it" is not done, even if there ist a done task
	q, err := CompileQQL("exists arg0 in items: testMacro(it) && done(arg0)")
	assert.Nil(t, err)

	t1 := todotxt.MustBuildItem(todotxt.WithDescription("T1"), todotxt.WithDone(true))
	t2 := todotxt.MustBuildItem(todotxt.WithDescription("T2"), todotxt.WithDone(false))
	testList := todotxt.ListOf(t1, t2)
	assert.True(t, q(testList, t2))
}
