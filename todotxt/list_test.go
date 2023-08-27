package todotxt

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestList_WritePlusReadIsTheIdentity(t *testing.T) {
	testCases := map[string]struct {
		itemList List
	}{
		"An empty list": {
			itemList: List{},
		},
		"A list with items": {
			itemList: List([]*Item{
				MustBuildItem(WithDescription("Hello World")),
				MustBuildItem(WithDescription("Hello World2")),
			}),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := bytes.Buffer{}
			tc.itemList.Save(&out, DefaultFormatter)
			listOut, err := Read(bytes.NewReader(out.Bytes()))
			assert.Nil(t, err)
			assert.Equal(t, tc.itemList, listOut)
		})
	}
}

func Test_ReadReturnsErrorsForInvalidItems(t *testing.T) {
	listString := `x (A) a done task with prio set
x 2022-13-01 a task with a malformed creation date
2022-10-01 2020-10-01 a task with completion date set although not done`

	_, err := Read(strings.NewReader(listString))

	sErr := err.(interface{ Unwrap() []error }).Unwrap()
	assert.Len(t, sErr, 3)
	var first, second, third ReadError
	assert.ErrorAs(t, sErr[0], &first)
	assert.ErrorAs(t, sErr[1], &second)
	assert.ErrorAs(t, sErr[2], &third)

	assert.Equal(t, 1, first.LineNumber)
	assert.Equal(t, "x (A) a done task with prio set", first.Line)
	assert.ErrorIs(t, first.BaseError, ErrNoPrioWhenDone)

	assert.Equal(t, 2, second.LineNumber)
	assert.Equal(t, "x 2022-13-01 a task with a malformed creation date", second.Line)
	assert.ErrorAs(t, second.BaseError, &ParseError{})

	assert.Equal(t, 3, third.LineNumber)
	assert.Equal(t, "2022-10-01 2020-10-01 a task with completion date set although not done", third.Line)
	assert.ErrorIs(t, third.BaseError, ErrCompletionDateWhileUndone)
}
