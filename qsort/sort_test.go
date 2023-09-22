package qsort_test

import (
	"testing"
	"time"

	"github.com/Fabian-G/quest/qsort"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

const (
	firstSmaller  = -1
	secondSmaller = 1
	same          = 0
)

func Test_Sort(t *testing.T) {
	testCases := map[string]struct {
		sortString    string
		first         *todotxt.Item
		second        *todotxt.Item
		expectedOrder int
	}{
		"sorting by done asc": {
			sortString:    "+done",
			first:         todotxt.MustBuildItem(todotxt.WithDone(false)),
			second:        todotxt.MustBuildItem(todotxt.WithDone(true)),
			expectedOrder: firstSmaller,
		},
		"sorting by done desc": {
			sortString:    "-done",
			first:         todotxt.MustBuildItem(todotxt.WithDone(false)),
			second:        todotxt.MustBuildItem(todotxt.WithDone(true)),
			expectedOrder: secondSmaller,
		},
		"sorting by priority asc": {
			sortString:    "+priority",
			first:         todotxt.MustBuildItem(todotxt.WithPriority(todotxt.PrioA)),
			second:        todotxt.MustBuildItem(todotxt.WithPriority(todotxt.PrioB)),
			expectedOrder: secondSmaller,
		},
		"sorting by priority desc": {
			sortString:    "-priority",
			first:         todotxt.MustBuildItem(todotxt.WithPriority(todotxt.PrioA)),
			second:        todotxt.MustBuildItem(todotxt.WithPriority(todotxt.PrioB)),
			expectedOrder: firstSmaller,
		},
		"no priority is the smallest": {
			sortString:    "+priority",
			first:         todotxt.MustBuildItem(todotxt.WithPriority(todotxt.PrioNone)),
			second:        todotxt.MustBuildItem(todotxt.WithPriority(todotxt.PrioZ)),
			expectedOrder: firstSmaller,
		},
		"sorting by creation asc": {
			sortString:    "+creation",
			first:         todotxt.MustBuildItem(todotxt.WithCreationDate(time.Now())),
			second:        todotxt.MustBuildItem(todotxt.WithCreationDate(time.Now().Add(24 * time.Hour))),
			expectedOrder: firstSmaller,
		},
		"sorting by creation desc": {
			sortString:    "-creation",
			first:         todotxt.MustBuildItem(todotxt.WithCreationDate(time.Now())),
			second:        todotxt.MustBuildItem(todotxt.WithCreationDate(time.Now().Add(24 * time.Hour))),
			expectedOrder: secondSmaller,
		},
		"sorting by completion asc": {
			sortString:    "+completion",
			first:         todotxt.MustBuildItem(todotxt.WithDone(true), todotxt.WithCompletionDate(time.Now()), todotxt.WithCreationDate(time.Now())),
			second:        todotxt.MustBuildItem(todotxt.WithDone(true), todotxt.WithCompletionDate(time.Now().Add(24*time.Hour)), todotxt.WithCreationDate(time.Now())),
			expectedOrder: firstSmaller,
		},
		"sorting by completion desc": {
			sortString:    "-completion",
			first:         todotxt.MustBuildItem(todotxt.WithDone(true), todotxt.WithCompletionDate(time.Now()), todotxt.WithCreationDate(time.Now())),
			second:        todotxt.MustBuildItem(todotxt.WithDone(true), todotxt.WithCompletionDate(time.Now().Add(24*time.Hour)), todotxt.WithCreationDate(time.Now())),
			expectedOrder: secondSmaller,
		},
		"null dates are always smaller": {
			sortString:    "+creation",
			first:         todotxt.MustBuildItem(todotxt.WithCreationDate(time.Now())),
			second:        todotxt.MustBuildItem(todotxt.WithoutCreationDate()),
			expectedOrder: secondSmaller,
		},
		"sort by description": {
			sortString:    "+description",
			first:         todotxt.MustBuildItem(todotxt.WithDescription("ABC")),
			second:        todotxt.MustBuildItem(todotxt.WithDescription("BCA")),
			expectedOrder: firstSmaller,
		},
		"sort by description desc": {
			sortString:    "-description",
			first:         todotxt.MustBuildItem(todotxt.WithDescription("ABC")),
			second:        todotxt.MustBuildItem(todotxt.WithDescription("BCA")),
			expectedOrder: secondSmaller,
		},
		"sort by tag asc": {
			sortString:    "+tag:hello",
			first:         todotxt.MustBuildItem(todotxt.WithDescription("hello:ABC")),
			second:        todotxt.MustBuildItem(todotxt.WithDescription("hello:BCA")),
			expectedOrder: firstSmaller,
		},
		"sort by tag desc": {
			sortString:    "-tag:hello",
			first:         todotxt.MustBuildItem(todotxt.WithDescription("hello:ABC")),
			second:        todotxt.MustBuildItem(todotxt.WithDescription("hello:BCA")),
			expectedOrder: secondSmaller,
		},
		"not present tags sort at the end": {
			sortString:    "+tag:hello",
			first:         todotxt.MustBuildItem(todotxt.WithDescription("hello:ABC")),
			second:        todotxt.MustBuildItem(todotxt.WithDescription("No tags here")),
			expectedOrder: secondSmaller,
		},
		"not present tags sort at the end (2)": {
			sortString:    "-tag:hello",
			first:         todotxt.MustBuildItem(todotxt.WithDescription("hello:ABC")),
			second:        todotxt.MustBuildItem(todotxt.WithDescription("No tags here")),
			expectedOrder: secondSmaller,
		},
		"sorting by multiple keys": {
			sortString:    "+done, -tag:hello, +tag:foo, -description",
			first:         todotxt.MustBuildItem(todotxt.WithDone(false), todotxt.WithDescription("A hello:world foo:abc")),
			second:        todotxt.MustBuildItem(todotxt.WithDone(false), todotxt.WithDescription("B hello:world foo:abc")),
			expectedOrder: secondSmaller,
		},
		"empty sort is always equal": {
			sortString:    "",
			first:         todotxt.MustBuildItem(todotxt.WithDone(false), todotxt.WithDescription("A hello:world foo:abc")),
			second:        todotxt.MustBuildItem(todotxt.WithDone(false), todotxt.WithDescription("B hello:world foo:abc")),
			expectedOrder: same,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cmpFunc, err := qsort.CompileSortFunc(tc.sortString, nil)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedOrder, cmpFunc(tc.first, tc.second))
		})
	}
}
