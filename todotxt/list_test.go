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
				MustBuild(WithDescription("Hello World")),
				MustBuild(WithDescription("Hello World2")),
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
	listString := `
x (A) a done task with prio set
x 2022-13-01 a task with a malformed creation date
2022-10-01 2020-10-01 a task with completion date set although not done
`

	_, err := Read(strings.NewReader(listString))
	assert.ErrorIs(t, err, ErrCompletionDateWhileUndone)
	assert.ErrorIs(t, err, ErrNoPrioWhenDone)
	assert.ErrorAs(t, err, &ParseError{})
}
