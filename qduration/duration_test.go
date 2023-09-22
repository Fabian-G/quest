package qduration_test

import (
	"testing"

	"github.com/Fabian-G/quest/qduration"
	"github.com/stretchr/testify/assert"
)

func Test_DurationParse(t *testing.T) {
	testCases := map[string]struct {
		durationString string
		duration       qduration.Duration
	}{
		"no span": {
			durationString: "+w",
			duration:       qduration.New(1, qduration.Week),
		},
		"negative span": {
			durationString: "-w",
			duration:       qduration.New(-1, qduration.Week),
		},
		"long span": {
			durationString: "100years",
			duration:       qduration.New(100, qduration.Year),
		},
		"negatve long span": {
			durationString: "-100years",
			duration:       qduration.New(-100, qduration.Year),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			d, err := qduration.Parse(tc.durationString)
			assert.Nil(t, err)
			assert.Equal(t, tc.duration, d)
		})
	}
}
