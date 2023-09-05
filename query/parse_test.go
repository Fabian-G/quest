package query

import (
	"strings"
	"testing"

	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/stretchr/testify/assert"
)

func Test_ParseConstructTheTreeCorrectly(t *testing.T) {
	testCases := map[string]struct {
		query               string
		expectedParseResult string
	}{
		"Single implication with constants": {
			query:               "true -> true",
			expectedParseResult: "(true -> true)",
		},
		"Single implication with quantor": {
			query:               "exists x: true -> true",
			expectedParseResult: "(exists x: (true -> true))",
		},
		"Deeply nested alternating quantors": {
			query:               "exists x: forall y: exists z: forall w: true -> true",
			expectedParseResult: "(exists x: (forall y: (exists z: (forall w: (true -> true)))))",
		},
		"And operation": {
			query:               "true && false",
			expectedParseResult: "(true && false)",
		},
		"Or operation": {
			query:               "true || false",
			expectedParseResult: "(true || false)",
		},
		"Not operation": {
			query:               "!true",
			expectedParseResult: "!true",
		},
		"Operator precedence": {
			query:               "!true && true || true -> true || true && !true",
			expectedParseResult: "(((!true && true) || true) -> (true || (true && !true)))",
		},
		"Using parans": {
			query:               "(false -> true) && (true -> (false -> true))",
			expectedParseResult: "((false -> true) && (true -> (false -> true)))",
		},
		"With function symbol": {
			query:               "done(it)",
			expectedParseResult: "done(it)",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			root, err := parseTree(tc.query, idSet{"it": struct{}{}})
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedParseResult, root.String())
		})
	}
}

func Test_eval(t *testing.T) {
	testCases := map[string]struct {
		list       *todotxt.List
		itemNumber int
		query      string
		result     bool
	}{
		"item is done (true)": {
			list: listFromString(t, `
			A not done item
			x a done item
			`),
			itemNumber: 0,
			query:      "done(it)",
			result:     false,
		},
		"item is done (false)": {
			list: listFromString(t, `
			A not done item
			x a done item
			`),
			itemNumber: 1,
			query:      "done(it)",
			result:     true,
		},
		"item has project +newKitchen (true)": {
			list: listFromString(t, `
			x an item with the +newKitchen Project
			an item without the newKitchen Project
			`),
			itemNumber: 0,
			query:      `contains(projects(it), "+newKitchen")`,
			result:     true,
		},
		"item has project +newKitchen (false)": {
			list: listFromString(t, `
			x an item with the +newKitchen Project
			an item without the newKitchen Project
			`),
			itemNumber: 1,
			query:      `contains(projects(it), "+newKitchen")`,
			result:     false,
		},
		"if all items of project +foo are done then it exists an item in project +bar that is not done (true)": {
			list: listFromString(t, `
			x a +foo item
			an item from +bar
			x another +foo item
			`),
			query:  `(forall f: contains(projects(f), "+foo") -> done(f)) -> (exists f: contains(projects(f), "+bar") && !done(f))`,
			result: true,
		},
		"if all items of project +foo are done then it exists an item in project +bar that is not done (false)": {
			list: listFromString(t, `
			x a +foo item
			x an item from +bar
			x another +foo item
			`),
			query:  `(forall f: contains(projects(f), "+foo") -> done(f)) -> (exists f: contains(projects(f), "+bar") && !done(f))`,
			result: false,
		},
		"all items are done (true)": {
			list: listFromString(t, `
			x a +foo item
			x an item from +bar
			x another +foo item
			`),
			query:  `forall t: done(t)`,
			result: true,
		},
		"all items are done (false)": {
			list: listFromString(t, `
			x a +foo item
			an item from +bar
			x another +foo item
			`),
			query:  `forall t: done(t)`,
			result: false,
		},
		"every item that is done is in project +foo (true)": {
			list: listFromString(t, `
			x a +foo item
			an item from +bar
			x another +foo item that is also in +bar
			`),
			query:  `forall i: done(i) -> contains(projects(i), "+foo")`,
			result: true,
		},
		"every item that is done is in project +foo (false)": {
			list: listFromString(t, `
			x a +foo item
			x an item from +bar
			another +foo item
			`),
			query:  `forall i: done(i) -> contains(projects(i), "+foo")`,
			result: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			queryFn, err := Compile(tc.query, FOL)
			assert.Nil(t, err)
			assert.Equal(t, tc.result, queryFn(tc.list, tc.list.Get(tc.itemNumber)))
		})
	}
}

func Test_InvalidQuerysResultInParseError(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"wrong type for and": {
			query: `true && "true"`,
		},
		"wrong type or and": {
			query: `true || "true"`,
		},
		"wrong type impl and": {
			query: `true -> "true"`,
		},
		"wrong type not and": {
			query: `!"true"`,
		},
		"wrong type for query": {
			query: `"hello world"`,
		},
		"unknown identifier": {
			query: `exists x: done(y)`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := Compile(tc.query, FOL)
			assert.Error(t, err)
		})
	}
}

func listFromString(t *testing.T, list string) *todotxt.List {
	tabsRemoved := strings.ReplaceAll(list, "\t", "")
	l, err := todotxt.DefaultDecoder.Decode(strings.NewReader(strings.TrimSpace(tabsRemoved)))
	assert.Nil(t, err)
	return todotxt.ListOf(l...)
}
