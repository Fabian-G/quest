package todotxt_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/stretchr/testify/assert"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestParse(t *testing.T) {
	testCases := map[string]struct {
		line          string
		expectedError error
		expectedItem  *todotxt.Item
	}{
		"Empty Todo Item": {
			line:         "\n", // Needs a trailing newline. Otherwise this would just result in an empty list
			expectedItem: todotxt.MustBuildItem(),
		},
		"Item with only a description": {
			line:         "This is a description",
			expectedItem: todotxt.MustBuildItem(todotxt.WithDescription("This is a description")),
		},
		"Item marked as done": {
			line:         "x a done item",
			expectedItem: todotxt.MustBuildItem(todotxt.WithDescription("a done item"), todotxt.WithDone(true)),
		},
		"Item with empty description": {
			line:         "x",
			expectedItem: todotxt.MustBuildItem(todotxt.WithDone(true)),
		},
		"Item with priority": {
			line:         "(D) an item with prio",
			expectedItem: todotxt.MustBuildItem(todotxt.WithPriority(todotxt.PrioD), todotxt.WithDescription("an item with prio")),
		},
		"Item with priority and empty description": {
			line:         "(D)",
			expectedItem: todotxt.MustBuildItem(todotxt.WithPriority(todotxt.PrioD)),
		},
		"A done item without completion date": {
			line: "x 2022-02-02 A done item",
			expectedItem: todotxt.MustBuildItem(
				todotxt.WithDone(true),
				todotxt.WithCreationDate(date(2022, 2, 2)),
				todotxt.WithDescription("A done item"),
			),
		},
		"A full task item": {
			line: "x 2022-02-02 2020-03-04 A +full @item",
			expectedItem: todotxt.MustBuildItem(
				todotxt.WithDone(true),
				todotxt.WithCompletionDate(date(2022, 2, 2)),
				todotxt.WithCreationDate(date(2020, 3, 4)),
				todotxt.WithDescription("A +full @item"),
			),
		},
		"A task containing a newline in the description": {
			line:         "This is a task\\nWith a new line in it",
			expectedItem: todotxt.MustBuildItem(todotxt.WithDescription("This is a task\nWith a new line in it")),
		},
		"A task with an invalid date produces a parser error": {
			line:          "2022-13-12 A task",
			expectedError: &todotxt.ParseError{},
		},
		"A task with an invalid priority is treated as description": {
			line:         "(?) A task",
			expectedItem: todotxt.MustBuildItem(todotxt.WithDescription("(?) A task")),
		},
		"A task starting with x, but without space is treated as description": {
			line:         "xTask",
			expectedItem: todotxt.MustBuildItem(todotxt.WithDescription("xTask")),
		},
		"Too much whitespace is ignored": {
			line: "x         2022-02-02     2020-03-04     A +full @item     ",
			expectedItem: todotxt.MustBuildItem(
				todotxt.WithDone(true),
				todotxt.WithCompletionDate(date(2022, 2, 2)),
				todotxt.WithCreationDate(date(2020, 3, 4)),
				todotxt.WithDescription("A +full @item"),
			),
		},
		"A task starting with whitespace is treated as description": {
			line: " x (F) 2022-02-02 2020-03-04 A +full @item",
			expectedItem: todotxt.MustBuildItem(
				todotxt.WithDescription("x (F) 2022-02-02 2020-03-04 A +full @item"),
			),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			itemList, err := todotxt.DefaultDecoder.Decode(strings.NewReader(tc.line))
			if e, ok := tc.expectedError.(*todotxt.ParseError); ok {
				assert.ErrorAs(t, err, e)
			} else if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.Equal(t, 1, itemList.Len())
				todotxt.AssertItemEqual(t, tc.expectedItem, itemList.Get(0))
			}
		})
	}
}

func Test_WellFormattedItemsShouldNotChangeAfterParsingPlusSerializing(t *testing.T) {
	testCases := map[string]struct {
		line string
	}{
		"Item with empty description": {
			line: "x",
		},
		"Item with only a description": {
			line: "This is a description",
		},
		"Item marked as done": {
			line: "x a done item",
		},
		"Item with priority": {
			line: "(D) an item with prio",
		},
		"A done item without completion date": {
			line: "x 2022-02-02 A done item",
		},
		"A full task item": {
			line: "x 2022-02-02 2020-03-04 A +full @item",
		},
		"A task starting with whitespace is treated as description": {
			line: " x (F) 2022-02-02 2020-03-04 A +full @item",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			itemList, err := todotxt.DefaultDecoder.Decode(strings.NewReader(tc.line))
			assert.Nil(t, err)
			assert.Equal(t, 1, itemList.Len())
			out := strings.Builder{}
			err = todotxt.DefaultEncoder.Encode(&out, itemList)
			assert.Nil(t, err)
			assert.Equal(t, tc.line, out.String())
		})
	}
}
