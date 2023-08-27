package todotxt_test

import (
	"testing"
	"time"

	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/stretchr/testify/assert"
)

func Test_DefaultFormat(t *testing.T) {
	creationDate := time.Date(2020, 4, 5, 6, 7, 8, 9, time.UTC)
	completionDate := time.Date(2023, 4, 5, 7, 8, 9, 10, time.UTC)
	dummy := todotxt.MustBuildItem(todotxt.WithMeta(true, todotxt.PrioNone, completionDate, creationDate), todotxt.WithDescription("This is a dummy task"))
	testCases := map[string]struct {
		item           *todotxt.Item
		expectedFormat string
	}{
		"A task with nothing but a description": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("this is a test")),
			expectedFormat: "this is a test",
		},
		"A full dummy task": {
			item:           todotxt.MustBuildItem(todotxt.CopyOf(dummy)),
			expectedFormat: "x 2023-04-05 2020-04-05 This is a dummy task",
		},
		"An uncompleted task": {
			item:           todotxt.MustBuildItem(todotxt.CopyOf(dummy), todotxt.WithDone(false), todotxt.WithCompletionDate(nil), todotxt.WithPriority(todotxt.PrioF)),
			expectedFormat: "(F) 2020-04-05 This is a dummy task",
		},
		"A task without priority": {
			item:           todotxt.MustBuildItem(todotxt.CopyOf(dummy), todotxt.WithDone(true)),
			expectedFormat: "x 2023-04-05 2020-04-05 This is a dummy task",
		},
		"A task without any dates": {
			item:           todotxt.MustBuildItem(todotxt.CopyOf(dummy), todotxt.WithCompletionDate(nil), todotxt.WithCreationDate(nil)),
			expectedFormat: "x This is a dummy task",
		},
		"Description with x in the beginning should start with space": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("x test")),
			expectedFormat: " x test",
		},
		"Description with date in the beginning should start with space": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("2012-03-04 test")),
			expectedFormat: " 2012-03-04 test",
		},
		"Description with Prio in the beginning should start with space": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("(A) test")),
			expectedFormat: " (A) test",
		},
		"Description with x in the beginning, but without space should not start with space": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("xTest")),
			expectedFormat: "xTest",
		},
		"Description with date in the beginning, but without space should not start with space": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("2012-03-04Test")),
			expectedFormat: "2012-03-04Test",
		},
		"Description with Prio in the beginning, but without space should not start with space": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("(A)test")),
			expectedFormat: "(A)test",
		},
		"Description with newline": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("A description\nSpanning\nMultiple Lines")),
			expectedFormat: "A description\\nSpanning\\nMultiple Lines",
		},
		"Description with windows style new line": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("A description\r\nSpanning\r\nMultiple Lines")),
			expectedFormat: "A description\\r\\nSpanning\\r\\nMultiple Lines",
		},
		"A description with trailing whitespace": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("A description with trailing space      \t   ")),
			expectedFormat: "A description with trailing space",
		},
		"A description with leading whitespace": {
			item:           todotxt.MustBuildItem(todotxt.WithDescription("   \t   A description with trailing space")),
			expectedFormat: "A description with trailing space",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out, err := todotxt.DefaultFormatter.Format(tc.item)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedFormat, out)
		})
	}
}

func Test_FormatReturnsErrorOnInvalidTask(t *testing.T) {
	item, _ := todotxt.BuildItem(todotxt.WithCompletionDate(new(time.Time)), todotxt.WithCreationDate(nil))

	_, err := todotxt.DefaultFormatter.Format(item)
	assert.Error(t, err)
}
