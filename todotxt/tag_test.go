package todotxt_test

import (
	"testing"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

func TestTag_MatcherForTag(t *testing.T) {
	testCases := map[string]struct {
		desc            string
		key             string
		expectedMatches [][]int
		expectedValues  []string
	}{
		"Matcher finds all matches": {
			desc: "tag:foo foo tag:bar bar tag:baz",
			key:  "tag",
			expectedMatches: [][]int{
				{0, 8}, {11, 20}, {23, 31},
			},
			expectedValues: []string{"foo", "bar", "baz"},
		},
		"Matcher finds only matches": {
			desc:            "tag2:foo foo :proj bar tag:",
			key:             "tag",
			expectedMatches: [][]int{},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			matcher := todotxt.MatcherForTag(tc.key)
			results := matcher.FindAllStringIndex(tc.desc, -1)
			assert.ElementsMatch(t, tc.expectedMatches, results)
			submatch := matcher.FindAllStringSubmatch(tc.desc, -1)
			values := make([]string, 0, len(tc.expectedValues))
			for _, match := range submatch {
				values = append(values, match[1])
			}
			assert.ElementsMatch(t, tc.expectedValues, values)
		})
	}
}

func Test_MatcherCanBeUsedToSetTagValue(t *testing.T) {
	description := "This is a test rec:+5y description rec:+6y with a tag"

	matcher := todotxt.MatcherForTag("rec")
	newDescription := matcher.ReplaceAllString(description, " rec:+1d ")

	assert.Equal(t, "This is a test rec:+1d description rec:+1d with a tag", newDescription)
}
