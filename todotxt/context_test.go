package todotxt_test

import (
	"testing"

	"github.com/Fabian-G/todotxt/todotxt"
	"github.com/stretchr/testify/assert"
)

func TestContext_Matcher(t *testing.T) {
	testCases := map[string]struct {
		desc            string
		context         todotxt.Context
		expectedMatches [][]int
	}{
		"Matcher finds all matches": {
			desc:    "@ctx foo @ctx bar @ctx",
			context: todotxt.Context("ctx"),
			expectedMatches: [][]int{
				{0, 5}, {8, 14}, {17, 22},
			},
		},
		"Matcher finds only matches": {
			desc:            "+ctx2 foo@ctx bar ctx@",
			context:         todotxt.Context("ctx"),
			expectedMatches: [][]int{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			matcher := tc.context.Matcher()
			results := matcher.FindAllStringIndex(tc.desc, -1)
			assert.ElementsMatch(t, tc.expectedMatches, results)
		})
	}
}
