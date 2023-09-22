package todotxt_test

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Fabian-G/quest/todotxt"
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
			item:           todotxt.MustBuildItem(todotxt.CopyOf(dummy), todotxt.WithDone(false), todotxt.WithoutCompletionDate(), todotxt.WithPriority(todotxt.PrioF)),
			expectedFormat: "(F) 2020-04-05 This is a dummy task",
		},
		"A task without priority": {
			item:           todotxt.MustBuildItem(todotxt.CopyOf(dummy), todotxt.WithDone(true)),
			expectedFormat: "x 2023-04-05 2020-04-05 This is a dummy task",
		},
		"A task without any dates": {
			item:           todotxt.MustBuildItem(todotxt.CopyOf(dummy), todotxt.WithoutCompletionDate(), todotxt.WithoutCreationDate()),
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
			list := todotxt.List{}
			list.Add(tc.item)
			out := strings.Builder{}
			err := todotxt.DefaultEncoder.Encode(&out, list.Tasks())
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedFormat+"\n", out.String())
		})
	}
}

func Test_FormatReturnsErrorOnInvalidTask(t *testing.T) {
	item, _ := todotxt.BuildItem(todotxt.WithCompletionDate(time.Now()), todotxt.WithoutCreationDate())

	err := todotxt.DefaultEncoder.Encode(io.Discard, []*todotxt.Item{item})
	assert.Error(t, err)
}

func TestList_WritePlusReadIsTheIdentity(t *testing.T) {
	testCases := map[string]struct {
		itemList *todotxt.List
	}{
		"An empty list": {
			itemList: &todotxt.List{},
		},
		"A list with items": {
			itemList: todotxt.ListOf(
				todotxt.MustBuildItem(todotxt.WithDescription("Hello World")),
				todotxt.MustBuildItem(todotxt.WithDescription("Hello World2")),
			),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := bytes.Buffer{}
			todotxt.DefaultEncoder.Encode(&out, tc.itemList.Tasks())
			listOut, err := todotxt.DefaultDecoder.Decode(bytes.NewReader(out.Bytes()))
			assert.Nil(t, err)
			todotxt.AssertListEqual(t, tc.itemList, todotxt.ListOf(listOut...))
		})
	}
}

func Test_ReadReturnsErrorsForInvalidItems(t *testing.T) {
	listString := `x (A) a done task with prio set
x 2022-13-01 a task with a malformed creation date
2022-10-01 2020-10-01 a task with completion date set although not done`

	_, err := todotxt.DefaultDecoder.Decode(strings.NewReader(listString))

	sErr := err.(interface{ Unwrap() []error }).Unwrap()
	assert.Len(t, sErr, 3)
	var first, second, third todotxt.ReadError
	assert.ErrorAs(t, sErr[0], &first)
	assert.ErrorAs(t, sErr[1], &second)
	assert.ErrorAs(t, sErr[2], &third)

	assert.Equal(t, 1, first.LineNumber)
	assert.Equal(t, "x (A) a done task with prio set", first.Line)
	assert.ErrorIs(t, first.BaseError, todotxt.ErrNoPrioWhenDone)

	assert.Equal(t, 2, second.LineNumber)
	assert.Equal(t, "x 2022-13-01 a task with a malformed creation date", second.Line)
	assert.ErrorAs(t, second.BaseError, &todotxt.ParseError{})

	assert.Equal(t, 3, third.LineNumber)
	assert.Equal(t, "2022-10-01 2020-10-01 a task with completion date set although not done", third.Line)
	assert.ErrorIs(t, third.BaseError, todotxt.ErrCompletionDateWhileUndone)
}
