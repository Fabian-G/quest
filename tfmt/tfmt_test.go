package tfmt_test

import (
	"testing"
	"time"

	item_factory "github.com/Fabian-G/todotxt/test"
	"github.com/Fabian-G/todotxt/tfmt"
	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/stretchr/testify/assert"
)

func Test_DefaultFormat(t *testing.T) {
	creationDate := time.Date(2020, 4, 5, 6, 7, 8, 9, time.UTC)
	completionDate := time.Date(2023, 4, 5, 7, 8, 9, 10, time.UTC)
	dummy := item_factory.MustBuild(todotxt.WithMeta(true, todotxt.PrioF, completionDate, creationDate), todotxt.WithDescription("This is a dummy task"))
	testCases := map[string]struct {
		item           *todotxt.Item
		expectedFormat string
	}{
		"Empty task": {
			item:           &todotxt.Item{},
			expectedFormat: "",
		},
		"Empty description": {
			item:           item_factory.Apply(dummy, item_factory.WithEmptyDescription),
			expectedFormat: "",
		},
		"A task with nothing but a description": {
			item:           item_factory.EmptyItem(todotxt.WithDescription("this is a test")),
			expectedFormat: "this is a test",
		},
		"A full dummy task": {
			item:           item_factory.Apply(dummy),
			expectedFormat: "x (F) 2023-04-05 2020-04-05 This is a dummy task",
		},
		"An uncompleted task": {
			item:           item_factory.Apply(dummy, item_factory.Uncompleted),
			expectedFormat: "(F) 2020-04-05 This is a dummy task",
		},
		"A task without priority": {
			item:           item_factory.Apply(dummy, item_factory.WithoutPriority),
			expectedFormat: "x 2023-04-05 2020-04-05 This is a dummy task",
		},
		"A task without any dates": {
			item:           item_factory.Apply(dummy, item_factory.WithoutCompletionDate, item_factory.WithoutCreationDate),
			expectedFormat: "x (F) This is a dummy task",
		},
		"Description with x in the beginning should start with space": {
			item:           item_factory.EmptyItem(todotxt.WithDescription("x test")),
			expectedFormat: " x test",
		},
		"Description with date in the beginning should start with space": {
			item:           item_factory.EmptyItem(todotxt.WithDescription("2012-03-04 test")),
			expectedFormat: " 2012-03-04 test",
		},
		"Description with Prio in the beginning should start with space": {
			item:           item_factory.EmptyItem(todotxt.WithDescription("(A) test")),
			expectedFormat: " (A) test",
		},
		"Description with x in the beginning, but without space should not start with space": {
			item:           item_factory.EmptyItem(todotxt.WithDescription("xTest")),
			expectedFormat: "xTest",
		},
		"Description with date in the beginning, but without space should not start with space": {
			item:           item_factory.EmptyItem(todotxt.WithDescription("2012-03-04Test")),
			expectedFormat: "2012-03-04Test",
		},
		"Description with Prio in the beginning, but without space should not start with space": {
			item:           item_factory.EmptyItem(todotxt.WithDescription("(A)test")),
			expectedFormat: "(A)test",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedFormat, tfmt.DefaultFormatter.Format(tc.item))
		})
	}
}

func Test_FormatPanicsOnInvalidTask(t *testing.T) {
	item, _ := todotxt.Build(todotxt.WithCompletionDate(new(time.Time)), todotxt.WithCreationDate(nil))

	assert.Panics(t, func() {
		_ = tfmt.DefaultFormatter.Format(item)
	})
}
