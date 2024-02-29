package qscore_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/Fabian-G/quest/qduration"
	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

var today = time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)
var testCalculator = qscore.Calculator{
	UrgencyTags:  []qscore.UrgencyTag{{Tag: "due"}},
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
		"item has urgency in the middle": {
			item: todotxt.MustBuildItem(todotxt.WithDescription(dueDateInDays(testCalculator.UrgencyBegin/2, "Hello World"))),
			expectedScore: qscore.Score{
				Urgency:    5.5,
				Importance: 0,
			},
		},
		"item has importance in the middle": {
			item: todotxt.MustBuildItem(todotxt.WithDescription("Hello World"), todotxt.WithPriority(todotxt.PrioC)),
			expectedScore: qscore.Score{
				Urgency:    0,
				Importance: 5.5,
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

func Test_ScoreIsMarkedUrgentAppropriately(t *testing.T) {
	testScore := qscore.Score{
		Urgency: 10,
	}

	assert.True(t, testScore.IsUrgent())
	testScore = qscore.Score{
		Urgency: 0,
	}

	assert.False(t, testScore.IsUrgent())
}
func Test_ScoreIsMarkedImportantAppropriately(t *testing.T) {
	testScore := qscore.Score{
		Importance: 10,
	}

	assert.True(t, testScore.IsImportant())
	testScore = qscore.Score{
		Importance: 0,
	}

	assert.False(t, testScore.IsImportant())
}

func Test_ScoreFallsBackToDefaultUrgencyWhenDueIsUnset(t *testing.T) {
	defaultUrgency, err := qduration.Parse("5d")
	assert.NoError(t, err)
	testCalculator := qscore.Calculator{
		UrgencyTags:    []qscore.UrgencyTag{{Tag: "due"}},
		UrgencyBegin:   10,
		DefaultUrgency: defaultUrgency,
		MinPriority:    todotxt.PrioE,
		NowFunc: func() time.Time {
			return today
		},
	}
	item := todotxt.MustBuildItem(todotxt.WithDescription("Hello World"), todotxt.WithCreationDate(today))
	score := testCalculator.ScoreOf(item)
	expected := qscore.Score{
		Urgency:    5.5,
		Importance: 0,
	}
	assertApproximatelyEqual(t, expected, score)
}

func Test_ScoreUsesOffsetCorrectly(t *testing.T) {
	offset, err := qduration.Parse("4d")
	assert.NoError(t, err)
	testCalculator := qscore.Calculator{
		UrgencyTags:  []qscore.UrgencyTag{{Tag: "due", Offset: offset}},
		UrgencyBegin: 10,
		MinPriority:  todotxt.PrioE,
		NowFunc: func() time.Time {
			return today
		},
	}
	item := todotxt.MustBuildItem(todotxt.WithDescription(dueDateInDays(1, "Hello World")))
	score := testCalculator.ScoreOf(item)
	expected := qscore.Score{
		Urgency:    5.5,
		Importance: 0,
	}
	assertApproximatelyEqual(t, expected, score)
}

func Test_EmptyCalculatorReturnsZeroValue(t *testing.T) {
	item := todotxt.MustBuildItem(todotxt.WithPriority(todotxt.PrioA), todotxt.WithDescription("Test due:1990-09-09"))
	assert.Equal(t, qscore.Score{}, qscore.Calculator{}.ScoreOf(item))
}

func Test_PrioritiesAreConsideredImportantAccordingToABCDEMethod(t *testing.T) {
	item := todotxt.MustBuildItem(todotxt.WithDescription("test"), todotxt.WithPriority(todotxt.PrioC))
	assert.True(t, testCalculator.ScoreOf(item).IsImportant())
	item.PrioritizeAs(todotxt.PrioD)
	assert.False(t, testCalculator.ScoreOf(item).IsImportant())
}

func Test_AnAlmostDueTaskShouldYieldAHightUrgency(t *testing.T) {
	item := todotxt.MustBuildItem(todotxt.WithDescription(dueDateInDays(1, "Hello World")))
	score := testCalculator.ScoreOf(item)
	assert.GreaterOrEqual(t, score.Urgency, float32(9))
	assert.True(t, score.IsUrgent())
}

func assertApproximatelyEqual(t *testing.T, expected qscore.Score, actual qscore.Score) {
	epsilon := 0.01
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
