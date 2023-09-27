package qselect_test

import (
	"testing"

	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

func Test_Range(t *testing.T) {
	testList := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("T1")),
		todotxt.MustBuildItem(todotxt.WithDescription("T2")),
		todotxt.MustBuildItem(todotxt.WithDescription("T3")),
		todotxt.MustBuildItem(todotxt.WithDescription("T4")),
		todotxt.MustBuildItem(todotxt.WithDescription("T5")),
		todotxt.MustBuildItem(todotxt.WithDescription("T6")),
		todotxt.MustBuildItem(todotxt.WithDescription("T7")),
		todotxt.MustBuildItem(todotxt.WithDescription("T8")),
		todotxt.MustBuildItem(todotxt.WithDescription("T9")),
	)
	testCases := map[string]struct {
		rng             string
		expectedMatches []string
	}{
		"single line": {
			rng:             "3",
			expectedMatches: []string{"T3"},
		},
		"multiple lines": {
			rng:             "3, 5, 7",
			expectedMatches: []string{"T3", "T5", "T7"},
		},
		"single range": {
			rng:             "3-5",
			expectedMatches: []string{"T3", "T4", "T5"},
		},
		"multiple ranges": {
			rng:             "3-5,7-8",
			expectedMatches: []string{"T3", "T4", "T5", "T7", "T8"},
		},
		"ranges and singles mixed": {
			rng:             "1,3,4-5,9",
			expectedMatches: []string{"T1", "T3", "T4", "T5", "T9"},
		},
		"open ranges": {
			rng:             "-4,8-",
			expectedMatches: []string{"T1", "T2", "T3", "T4", "T8", "T9"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			query, err := qselect.CompileRange(tc.rng)
			assert.Nil(t, err)
			matches := query.Filter(testList)
			descriptions := make([]string, 0, len(matches))
			for _, t := range matches {
				descriptions = append(descriptions, t.Description())
			}
			assert.ElementsMatch(t, tc.expectedMatches, descriptions)
		})
	}
}
