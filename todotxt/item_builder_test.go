package todotxt_test

import (
	"testing"
	"time"

	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/stretchr/testify/assert"
)

func TestBuildErrors(t *testing.T) {
	dateOne := time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)
	dateAfterOne := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)
	testCases := map[string]struct {
		expectedError error
		buildConfig   []todotxt.BuildFunc
	}{
		"Task with completion date set, but not creation date": {
			expectedError: todotxt.ErrCreationDateUnset,
			buildConfig: []todotxt.BuildFunc{
				todotxt.WithDone(true),
				todotxt.WithDescription("Test"),
				todotxt.WithCreationDate(nil),
				todotxt.WithCompletionDate(&dateOne),
			},
		},
		"Task with completion date before creation date": {
			expectedError: todotxt.ErrCompleteBeforeCreation,
			buildConfig: []todotxt.BuildFunc{
				todotxt.WithDone(true),
				todotxt.WithDescription("Test"),
				todotxt.WithCreationDate(&dateAfterOne),
				todotxt.WithCompletionDate(&dateOne),
			},
		},
		"Undone task with completion date": {
			expectedError: todotxt.ErrCompletionDateWhileUndone,
			buildConfig: []todotxt.BuildFunc{
				todotxt.WithDone(false),
				todotxt.WithDescription("Test"),
				todotxt.WithCreationDate(&dateOne),
				todotxt.WithCompletionDate(&dateAfterOne),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := todotxt.BuildItem(tc.buildConfig...)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}
