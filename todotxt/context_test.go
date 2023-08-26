package todotxt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext_Matcher(t *testing.T) {
	testCases := map[string]struct {
		desc            string
		context         Context
		expectedMatches [][]int
	}{
		"Matcher finds all matches": {
			desc:    "@ctx foo @ctx bar @ctx",
			context: Context("ctx"),
			expectedMatches: [][]int{
				{0, 5}, {8, 14}, {17, 22},
			},
		},
		"Matcher finds only matches": {
			desc:            "+ctx2 foo@ctx bar ctx@",
			context:         Context("ctx"),
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

func TestContext_String(t *testing.T) {
	tests := []struct {
		name string
		c    Context
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("Context.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
