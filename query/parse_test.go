package query

import (
	"strings"
	"testing"

	"github.com/Fabian-G/quest/todotxt"
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
			query:               "exists x in items: true -> true",
			expectedParseResult: "(exists x in (items): (true -> true))",
		},
		"Deeply nested alternating quantors": {
			query:               "exists x in items: forall y in items: exists z in items: forall w in items: true -> true",
			expectedParseResult: "(exists x in (items): (forall y in (items): (exists z in (items): (forall w in (items): (true -> true)))))",
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
			query:               "!true == false && true || true -> true || true && !true == false",
			expectedParseResult: "(((!(true == false) && true) || true) -> (true || (true && !(true == false))))",
		},
		"Using parans": {
			query:               "(false -> true) && (true -> (false -> true))",
			expectedParseResult: "((false -> true) && (true -> (false -> true)))",
		},
		"With function symbol": {
			query:               "done(it)",
			expectedParseResult: "done(it)",
		},
		"high precedence followed by low precedence": {
			query:               "true && false || true",
			expectedParseResult: "((true && false) || true)",
		},
		"Quantor in the middle": {
			query:               "!done(it) && exists x in items: done(x)",
			expectedParseResult: "(!done(it) && (exists x in (items): done(x)))",
		},
		"it is optional for functions that require one item": {
			query:               "done()",
			expectedParseResult: "done(it)",
		},
		"empty parens can be omitted": {
			query:               "done",
			expectedParseResult: "done(it)",
		},
		"space is optional": {
			query:               "done||!done",
			expectedParseResult: "(done(it) || !done(it))",
		},
		"bound id with name of a function gets parsed as identifier": {
			query:               "exists done in items: done(done)",
			expectedParseResult: "(exists done in (items): done(done))",
		},
		"project match shorthand gets compiled correctly": {
			query:               "+foo",
			expectedParseResult: "(exists p in (projects(it)): dotPrefix(p, \"+foo\"))",
		},
		"context match shorthand gets compiled correctly": {
			query:               "@foo",
			expectedParseResult: "(exists p in (contexts(it)): dotPrefix(p, \"@foo\"))",
		},
		"stuff after project matcher": {
			query:               "+foo && !done",
			expectedParseResult: "((exists p in (projects(it)): dotPrefix(p, \"+foo\")) && !done(it))",
		},
		"chained and": {
			query:               "done && done && done",
			expectedParseResult: "((done(it) && done(it)) && done(it))",
		},
		"project matcher in between": {
			query:               "done && +foo && !done",
			expectedParseResult: "((done(it) && (exists p in (projects(it)): dotPrefix(p, \"+foo\"))) && !done(it))",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			root, err := parseTree(tc.query, idSet{"it": qItem, "items": qItemSlice})
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
			query:      `exists p in projects(it): p == "+newKitchen"`,
			result:     true,
		},
		"item has project +newKitchen (false)": {
			list: listFromString(t, `
			x an item with the +newKitchen Project
			an item without the newKitchen Project
			`),
			itemNumber: 1,
			query:      `exists p in projects(it): p == "+newKitchen"`,
			result:     false,
		},
		"if all items of project +foo are done then it exists an item in project +bar that is not done (true)": {
			list: listFromString(t, `
			x a +foo item
			an item from +bar
			x another +foo item
			`),
			query:  `(forall f in items: (exists p in projects(f): p == "+foo") -> done(f)) -> (exists f in items: (exists p in projects(f): p == "+bar") && !done(f))`,
			result: true,
		},
		"if all items of project +foo are done then it exists an item in project +bar that is not done (false)": {
			list: listFromString(t, `
			x a +foo item
			x an item from +bar
			x another +foo item
			`),
			query:  `(forall f in items: (exists p in projects(f): p == "+foo") -> done(f)) -> (exists f in items: (exists p in projects(f): p == "+bar") && !done(f))`,
			result: false,
		},
		"all items are done (true)": {
			list: listFromString(t, `
			x a +foo item
			x an item from +bar
			x another +foo item
			`),
			query:  `forall t in items: done(t)`,
			result: true,
		},
		"all items are done (false)": {
			list: listFromString(t, `
			x a +foo item
			an item from +bar
			x another +foo item
			`),
			query:  `forall t in items: done(t)`,
			result: false,
		},
		"every item that is done is in project +foo (true)": {
			list: listFromString(t, `
			x a +foo item
			an item from +bar
			x another +foo item that is also in +bar
			`),
			query:  `forall i in items: done(i) -> (exists p in projects(i): p == "+foo")`,
			result: true,
		},
		"every item that is done is in project +foo (false)": {
			list: listFromString(t, `
			x a +foo item
			x an item from +bar
			another +foo item
			`),
			query:  `forall i in items: done(i) -> (exists p in projects(i): p == "+foo")`,
			result: false,
		},
		"can bind function name as identifier": {
			list: listFromString(t, `
			x a +foo item
			x an item from +bar
			another +foo item
			`),
			query:  `forall done in items: done(done)`,
			result: false,
		},
		"quantifier can bind to to arbitrary slice typed expressions": {
			list: listFromString(t, `
			x a +foo item +bar
			x an item from +bar
			another +foo item
			`),
			query:  `forall i in items: forall proj in projects(i): exists other in items: !i == other && (exists p in projects(other): p == proj)`,
			result: true,
		},
		"project matcher matches project prefixes": {
			list: listFromString(t, `
			another +foo.bar item
			`),
			query:      `+foo`,
			result:     true,
			itemNumber: 0,
		},
		"project matcher matches project prefixes (false)": {
			list: listFromString(t, `
			another +foo2.bar item
			`),
			query:      `+foo`,
			result:     false,
			itemNumber: 0,
		},
		"context matcher matches context prefixes": {
			list: listFromString(t, `
			another +foo.bar item
			`),
			query:      `+foo`,
			result:     true,
			itemNumber: 0,
		},
		"context matcher matches context prefixes (false)": {
			list: listFromString(t, `
			another @foo2.bar item
			`),
			query:      `@foo`,
			result:     false,
			itemNumber: 0,
		},
		"project matcher does not match prefix without dot": {
			list: listFromString(t, `
			another +foo.bar item
			`),
			query:      `+fo`,
			result:     false,
			itemNumber: 0,
		},
		"empty query matches everything": {
			list: listFromString(t, `
			another +foo.bar item
			`),
			query:      ``,
			result:     true,
			itemNumber: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			queryFn, err := Compile(tc.query, QQL)
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
		"wrong type for or": {
			query: `true || "true"`,
		},
		"wrong type for impl": {
			query: `true -> "true"`,
		},
		"wrong type for not": {
			query: `!"true"`,
		},
		"wrong type for query": {
			query: `"hello world"`,
		},
		"unknown identifier": {
			query: `exists x in items: done(y)`,
		},
		"wrong collection type": {
			query: `exists x in done: done(x)`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := Compile(tc.query, QQL)
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
