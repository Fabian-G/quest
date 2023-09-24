package qscore_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

var today = time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)
var testCalculator = qscore.Calculator{
	UrgencyTag:   "due",
	UrgencyBegin: 10,
	MinPriority:  todotxt.PrioE,
	NowFunc: func() time.Time {
		return today
	},
}

func Test_Calculator(t *testing.T) {
	testCases := map[string]struct {
		item          *todotxt.Item
		expectedScore qscore.Score
	}{
		"no due or priority": {
			item: todotxt.MustBuildItem(todotxt.WithDescription("Hello World")),
			expectedScore: qscore.Score{
				Urgency:    0.0,
				Importance: 0.0,
			},
		},
		"item is done": {
			item: todotxt.MustBuildItem(todotxt.WithDone(true), todotxt.WithDescription("Hello World due:1990-09-09")),
			expectedScore: qscore.Score{
				Urgency:    0.0,
				Importance: 0.0,
			},
		},
		"item has due date very far in the future": {
			item: todotxt.MustBuildItem(todotxt.WithDescription(dueDateInDays(testCalculator.UrgencyBegin, "Hello World"))),
			expectedScore: qscore.Score{
				Urgency:    1,
				Importance: 0.0,
			},
		},
		"item has very low priority": {
			item: todotxt.MustBuildItem(todotxt.WithDescription("Hello World"), todotxt.WithPriority(testCalculator.MinPriority)),
			expectedScore: qscore.Score{
				Urgency:    0.0,
				Importance: 1,
			},
		},
		"item has due date today": {
			item: todotxt.MustBuildItem(todotxt.WithDescription(dueDateInDays(0, "Hello World"))),
			expectedScore: qscore.Score{
				Urgency:    10,
				Importance: 0.0,
			},
		},
		"highest priority": {
			item: todotxt.MustBuildItem(todotxt.WithDescription("Hello World"), todotxt.WithPriority(todotxt.PrioA)),
			expectedScore: qscore.Score{
				Urgency:    0.0,
				Importance: 10,
			},
		},
		"item is urgent and important": {
			item: todotxt.MustBuildItem(todotxt.WithDescription(dueDateInDays(0, "Hello World")), todotxt.WithPriority(todotxt.PrioA)),
			expectedScore: qscore.Score{
				Urgency:    10,
				Importance: 10,
			},
		},
		"item has low urgency and importance": {
			item: todotxt.MustBuildItem(todotxt.WithDescription(dueDateInDays(testCalculator.UrgencyBegin, "Hello World")), todotxt.WithPriority(testCalculator.MinPriority)),
			expectedScore: qscore.Score{
				Urgency:    1,
				Importance: 1,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			score := testCalculator.ScoreOf(tc.item)
			assertApproximatelyEqual(t, tc.expectedScore, score)
		})
	}
}

func assertApproximatelyEqual(t *testing.T, expected qscore.Score, actual qscore.Score) {
	epsilon := 0.1
	urgencyDiff := math.Abs(float64(expected.Urgency - actual.Urgency))
	importanceDiff := math.Abs(float64(expected.Importance - actual.Importance))
	assert.LessOrEqual(t, urgencyDiff, epsilon, "urgency diverges more than epsilon")
	assert.LessOrEqual(t, importanceDiff, epsilon, "importance diverges more than epsilon")

	// sanity check the score
	assert.LessOrEqual(t, actual.Score, max(actual.Importance, actual.Urgency))
	assert.GreaterOrEqual(t, actual.Score, min(actual.Importance, actual.Urgency))
}

func dueDateInDays(days int, desc string) string {
	dueDate := today.Add(time.Duration(days) * 24 * time.Hour)
	return fmt.Sprintf("%s due:%s", desc, dueDate.Format(time.DateOnly))
}
