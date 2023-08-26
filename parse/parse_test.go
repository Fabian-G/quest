package parse_test

import (
	"testing"
	"time"

	"github.com/Fabian-G/todotxt/parse"
	"github.com/Fabian-G/todotxt/test"
	"github.com/Fabian-G/todotxt/tfmt"
	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/stretchr/testify/assert"
)

func datePtr(year int, month time.Month, day int) *time.Time {
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &date
}

func TestParse(t *testing.T) {
	testCases := map[string]struct {
		line          string
		expectedError error
		expectedItem  *todotxt.Item
	}{
		"Empty Todo Item": {
			line:          "",
			expectedError: parse.ErrEmpty,
		},
		"Item with only a description": {
			line:         "This is a description",
			expectedItem: test.MustBuild(todotxt.WithDescription("This is a description")),
		},
		"Item marked as done": {
			line:         "x a done item",
			expectedItem: test.MustBuild(todotxt.WithDescription("a done item"), todotxt.WithDone(true)),
		},
		"Item with empty description": {
			line:         "x",
			expectedItem: test.MustBuild(todotxt.WithDone(true)),
		},
		"Item with priority": {
			line:         "(D) an item with prio",
			expectedItem: test.MustBuild(todotxt.WithPriority(todotxt.PrioD), todotxt.WithDescription("an item with prio")),
		},
		"Item with priority and empty description": {
			line:         "(D)",
			expectedItem: test.MustBuild(todotxt.WithPriority(todotxt.PrioD)),
		},
		"A done item without completion date": {
			line: "x 2022-02-02 A done item",
			expectedItem: test.MustBuild(
				todotxt.WithDone(true),
				todotxt.WithCreationDate(datePtr(2022, 2, 2)),
				todotxt.WithDescription("A done item"),
			),
		},
		"A full task item": {
			line: "x (F) 2022-02-02 2020-03-04 A +full @item",
			expectedItem: test.MustBuild(
				todotxt.WithDone(true),
				todotxt.WithCompletionDate(datePtr(2022, 2, 2)),
				todotxt.WithCreationDate(datePtr(2020, 3, 4)),
				todotxt.WithDescription("A +full @item"),
				todotxt.WithPriority(todotxt.PrioF),
			),
		},
		"A task with an invalid date gets treated as description": {
			line:         "2022-13-12 A task",
			expectedItem: test.MustBuild(todotxt.WithDescription("2022-13-12 A task")),
		},
		"A task with an invalid priority gets treated as description": {
			line:         "(?) A task",
			expectedItem: test.MustBuild(todotxt.WithDescription("(?) A task")),
		},
		"A task starting with x, but without space is treated as description": {
			line:         "xTask",
			expectedItem: test.MustBuild(todotxt.WithDescription("xTask")),
		},
		"Too much whitespace is ignored": {
			line: "x     (F)    2022-02-02     2020-03-04     A +full @item     ",
			expectedItem: test.MustBuild(
				todotxt.WithDone(true),
				todotxt.WithCompletionDate(datePtr(2022, 2, 2)),
				todotxt.WithCreationDate(datePtr(2020, 3, 4)),
				todotxt.WithDescription("A +full @item"),
				todotxt.WithPriority(todotxt.PrioF),
			),
		},
		"A task starting with whitespace is treated as description": {
			line: " x (F) 2022-02-02 2020-03-04 A +full @item",
			expectedItem: test.MustBuild(
				todotxt.WithDescription("x (F) 2022-02-02 2020-03-04 A +full @item"),
			),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			item, err := parse.Item(tc.line)
			assert.ErrorIs(t, err, tc.expectedError)
			if tc.expectedError == nil {
				assert.Equal(t, tc.expectedItem, item)
			}
		})
	}
}

func Test_WellFormattedItemsShouldNotChangeAfterParsingPlusSerializing(t *testing.T) {
	testCases := map[string]struct {
		line         string
		expectedItem *todotxt.Item
	}{
		"Item with only a description": {
			line: "This is a description",
		},
		"Item marked as done": {
			line: "x a done item",
		},
		"Item with empty description": {
			line: "x",
		},
		"Item with priority": {
			line: "(D) an item with prio",
		},
		"Item with priority and empty description": {
			line: "(D)",
		},
		"A done item without completion date": {
			line: "x 2022-02-02 A done item",
		},
		"A full task item": {
			line: "x (F) 2022-02-02 2020-03-04 A +full @item",
		},
		"A task starting with whitespace is treated as description": {
			line: " x (F) 2022-02-02 2020-03-04 A +full @item",
			expectedItem: test.MustBuild(
				todotxt.WithDescription("x (F) 2022-02-02 2020-03-04 A +full @item"),
			),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			item, err := parse.Item(tc.line)
			assert.Nil(t, err)
			assert.Equal(t, tc.line, tfmt.DefaultFormatter.Format(item))
		})
	}
}
