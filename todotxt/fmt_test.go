package todotxt_test

import (
	"testing"

	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/stretchr/testify/assert"
)

func Test_DefaultFormat(t *testing.T) {
	testCases := map[string]struct {
		item           *todotxt.Item
		expectedFormat string
	}{
		"Empty task": {
			item:           &todotxt.Item{},
			expectedFormat: "",
		},
		"Empty description": {
			item:           todotxt.DummyItem(todotxt.WithEmptyDescription),
			expectedFormat: "",
		},
		"A task with nothing but a description": {
			item:           todotxt.EmptyItem(todotxt.WithDescription("this is a test")),
			expectedFormat: "this is a test",
		},
		"A full dummy task": {
			item:           todotxt.DummyItem(),
			expectedFormat: "x (F) 2023-04-05 2020-04-05 This is a dummy task",
		},
		"An uncompleted task": {
			item:           todotxt.DummyItem(todotxt.Uncompleted),
			expectedFormat: "(F) 2020-04-05 This is a dummy task",
		},
		"A task without priority": {
			item:           todotxt.DummyItem(todotxt.WithoutPriority),
			expectedFormat: "x 2023-04-05 2020-04-05 This is a dummy task",
		},
		"A task without any dates": {
			item:           todotxt.DummyItem(todotxt.WithoutCreationDate, todotxt.WithoutCompletionDate),
			expectedFormat: "x (F) This is a dummy task",
		},
		"Description with x in the beginning should start with space": {
			item:           todotxt.EmptyItem(todotxt.WithDescription("x test")),
			expectedFormat: " x test",
		},
		"Description with date in the beginning should start with space": {
			item:           todotxt.EmptyItem(todotxt.WithDescription("2012-03-04 test")),
			expectedFormat: " 2012-03-04 test",
		},
		"Description with Prio in the beginning should start with space": {
			item:           todotxt.EmptyItem(todotxt.WithDescription("(A) test")),
			expectedFormat: " (A) test",
		},
		"Description with x in the beginning, but without space should not start with space": {
			item:           todotxt.EmptyItem(todotxt.WithDescription("xTest")),
			expectedFormat: "xTest",
		},
		"Description with date in the beginning, but without space should not start with space": {
			item:           todotxt.EmptyItem(todotxt.WithDescription("2012-03-04Test")),
			expectedFormat: "2012-03-04Test",
		},
		"Description with Prio in the beginning, but without space should not start with space": {
			item:           todotxt.EmptyItem(todotxt.WithDescription("(A)test")),
			expectedFormat: "(A)test",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedFormat, todotxt.DefaultFormatter.Format(tc.item))
		})
	}
}
