package todotxt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject_Matcher(t *testing.T) {
	testCases := map[string]struct {
		desc            string
		proj            Project
		expectedMatches [][]int
	}{
		"Matcher finds all matches": {
			desc: "+proj foo +proj bar +proj",
			proj: Project("+proj"),
			expectedMatches: [][]int{
				{0, 6}, {9, 16}, {19, 25},
			},
		},
		"Matcher finds only matches": {
			desc:            "+proj2 foo+proj bar proj+",
			proj:            Project("+proj"),
			expectedMatches: [][]int{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			matcher := tc.proj.Matcher()
			results := matcher.FindAllStringIndex(tc.desc, -1)
			assert.ElementsMatch(t, tc.expectedMatches, results)
		})
	}
}
