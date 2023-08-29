package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseConstructTheTreeCorrectly(t *testing.T) {
	testCases := map[string]struct {
		query               string
		expectedParseResult string
	}{
		"Single implication with constants": {
			query:               "true -> true",
			expectedParseResult: "(true -> true)",
		},
		"Single implication with quantor": {
			query:               "exists x: true -> true",
			expectedParseResult: "(exists x: (true -> true))",
		},
		"Deeply nested alternating quantors": {
			query:               "exists x: forall y: exists z: forall w: true -> true",
			expectedParseResult: "(exists x: (forall y: (exists z: (forall w: (true -> true)))))",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			root, err := parseTree(tc.query)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedParseResult, root.String())
		})
	}
}
