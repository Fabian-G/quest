package hook_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Fabian-G/quest/hook"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

var defaultTags = hook.RecurrenceTags{
	Rec:       "rec",
	Due:       "due",
	Threshold: "t",
}

func Test_dueDateRecurrence(t *testing.T) {
	testCases := map[string]struct {
		dueDate         string
		recurrence      string
		expectedDueDate string
		today           string
	}{
		"day absolute recurrence": {
			dueDate:         "2023-02-27",
			recurrence:      "+3d",
			expectedDueDate: "2023-03-02",
		},
		"day absolute recurrence (alternate unit)": {
			dueDate:         "2023-08-01",
			recurrence:      "+3days",
			expectedDueDate: "2023-08-04",
		},
		"week absolute recurrence": {
			dueDate:         "2023-08-01",
			recurrence:      "+3w",
			expectedDueDate: "2023-08-22",
		},
		"week absolute recurrence (alternate unit)": {
			dueDate:         "2023-08-01",
			recurrence:      "+3weeks",
			expectedDueDate: "2023-08-22",
		},
		"month absolute recurrence": {
			dueDate:         "2023-08-01",
			recurrence:      "+3m",
			expectedDueDate: "2023-11-01",
		},
		"month absolute recurrence (alternate unit)": {
			dueDate:         "2023-08-01",
			recurrence:      "+3months",
			expectedDueDate: "2023-11-01",
		},
		"month recurrence over year boundary": {
			dueDate:         "2023-08-01",
			recurrence:      "+6months",
			expectedDueDate: "2024-02-01",
		},
		"year absolute recurrence": {
			dueDate:         "2023-08-01",
			recurrence:      "+3y",
			expectedDueDate: "2026-08-01",
		},
		"year absolute recurrence (alternate unit)": {
			dueDate:         "2023-08-01",
			recurrence:      "+3years",
			expectedDueDate: "2026-08-01",
		},
		"day relative recurrence": {
			dueDate:         "2023-08-01",
			recurrence:      "3d",
			today:           "2022-05-05",
			expectedDueDate: "2022-05-08",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var nowFunc func() time.Time
			if len(tc.today) > 0 {
				nowFunc = func() time.Time {
					t, _ := time.Parse(time.DateOnly, tc.today)
					return t
				}
			}
			recurrentItem := todotxt.MustBuildItem(
				todotxt.WithDescription(fmt.Sprintf("A recurrent item rec:%s due:%s", tc.recurrence, tc.dueDate)),
				todotxt.WithNowFunc(nowFunc),
			)
			list := todotxt.ListOf(recurrentItem)
			list.AddHook(hook.NewRecurrenceWithNowFunc(defaultTags, nowFunc))

			err := recurrentItem.Complete()

			assert.Nil(t, err)
			assert.Equal(t, 2, list.Len())
			assert.True(t, list.GetLine(1).Done())
			assert.Equal(t, fmt.Sprintf("A recurrent item rec:%s due:%s", tc.recurrence, tc.dueDate), list.GetLine(1).Description())
			assert.False(t, list.GetLine(2).Done())
			assert.Equal(t, fmt.Sprintf("A recurrent item rec:%s due:%s", tc.recurrence, tc.expectedDueDate), list.GetLine(2).Description())
		})
	}
}

func Test_thresholdDateRecurrence(t *testing.T) {
	testCases := map[string]struct {
		tDate         string
		recurrence    string
		expectedTDate string
		today         string
	}{
		"day absolute recurrence": {
			tDate:         "2023-08-01",
			recurrence:    "+3d",
			expectedTDate: "2023-08-04",
		},
		"day absolute recurrence (alternate unit)": {
			tDate:         "2023-08-01",
			recurrence:    "+3days",
			expectedTDate: "2023-08-04",
		},
		"week absolute recurrence": {
			tDate:         "2023-08-01",
			recurrence:    "+3w",
			expectedTDate: "2023-08-22",
		},
		"week absolute recurrence (alternate unit)": {
			tDate:         "2023-08-01",
			recurrence:    "+3weeks",
			expectedTDate: "2023-08-22",
		},
		"month absolute recurrence": {
			tDate:         "2023-08-01",
			recurrence:    "+3m",
			expectedTDate: "2023-11-01",
		},
		"month absolute recurrence (alternate unit)": {
			tDate:         "2023-08-01",
			recurrence:    "+3months",
			expectedTDate: "2023-11-01",
		},
		"year absolute recurrence": {
			tDate:         "2023-08-01",
			recurrence:    "+3y",
			expectedTDate: "2026-08-01",
		},
		"year absolute recurrence (alternate unit)": {
			tDate:         "2023-08-01",
			recurrence:    "+3years",
			expectedTDate: "2026-08-01",
		},
		"day relative recurrence": {
			tDate:         "2023-08-01",
			recurrence:    "3d",
			today:         "2022-05-05",
			expectedTDate: "2022-05-08",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var nowFunc func() time.Time
			if len(tc.today) > 0 {
				nowFunc = func() time.Time {
					t, _ := time.Parse(time.DateOnly, tc.today)
					return t
				}
			}
			recurrentItem := todotxt.MustBuildItem(
				todotxt.WithDescription(fmt.Sprintf("A recurrent item rec:%s t:%s", tc.recurrence, tc.tDate)),
				todotxt.WithNowFunc(nowFunc),
			)
			list := todotxt.ListOf(recurrentItem)
			list.AddHook(hook.NewRecurrenceWithNowFunc(defaultTags, nowFunc))

			err := recurrentItem.Complete()

			assert.Nil(t, err)
			assert.Equal(t, 2, list.Len())
			assert.True(t, list.GetLine(1).Done())
			assert.Equal(t, fmt.Sprintf("A recurrent item rec:%s t:%s", tc.recurrence, tc.tDate), list.GetLine(1).Description())
			assert.False(t, list.GetLine(2).Done())
			assert.Equal(t, fmt.Sprintf("A recurrent item rec:%s t:%s", tc.recurrence, tc.expectedTDate), list.GetLine(2).Description())
		})
	}
}

func Test_bothDateRecurrence(t *testing.T) {
	testCases := map[string]struct {
		dueDate         string
		tDate           string
		recurrence      string
		expectedTDate   string
		expectedDueDate string
		today           string
	}{
		"day absolute recurrence": {
			tDate:           "2023-08-01",
			dueDate:         "2023-08-06",
			recurrence:      "+3d",
			expectedTDate:   "2023-08-04",
			expectedDueDate: "2023-08-09",
		},
		"day relative recurrence": {
			tDate:           "2023-06-04",
			dueDate:         "2023-08-06",
			recurrence:      "3d",
			today:           "2023-07-09",
			expectedTDate:   "2023-07-12",
			expectedDueDate: "2023-09-13",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var nowFunc func() time.Time
			if len(tc.today) > 0 {
				nowFunc = func() time.Time {
					t, _ := time.Parse(time.DateOnly, tc.today)
					return t
				}
			}
			recurrentItem := todotxt.MustBuildItem(
				todotxt.WithDescription(fmt.Sprintf("A recurrent item rec:%s t:%s due:%s", tc.recurrence, tc.tDate, tc.dueDate)),
				todotxt.WithNowFunc(nowFunc),
			)
			list := todotxt.ListOf(recurrentItem)
			list.AddHook(hook.NewRecurrenceWithNowFunc(defaultTags, nowFunc))

			err := recurrentItem.Complete()

			assert.Nil(t, err)
			assert.Equal(t, 2, list.Len())
			assert.True(t, list.GetLine(1).Done())
			assert.Equal(t, fmt.Sprintf("A recurrent item rec:%s t:%s due:%s", tc.recurrence, tc.tDate, tc.dueDate), list.GetLine(1).Description())
			assert.False(t, list.GetLine(2).Done())
			assert.Equal(t, fmt.Sprintf("A recurrent item rec:%s t:%s due:%s", tc.recurrence, tc.expectedTDate, tc.expectedDueDate), list.GetLine(2).Description())
		})
	}
}

func Test_CreationDateGetsUpdatedToToday(t *testing.T) {
	recurrentItem := todotxt.MustBuildItem(
		todotxt.WithDescription("A recurrent item rec:+5d due:2023-01-01"),
	)
	list := todotxt.ListOf(recurrentItem)
	list.AddHook(hook.NewRecurrenceWithNowFunc(defaultTags, func() time.Time { return time.Date(1990, 5, 5, 10, 2, 3, 4, time.UTC) }))

	err := recurrentItem.Complete()

	assert.Nil(t, err)
	expectedDate := time.Date(1990, 5, 5, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, &expectedDate, list.GetLine(2).CreationDate())
	assert.Nil(t, list.GetLine(2).CompletionDate())
}

func Test_RecurrenceValidationOnMissing(t *testing.T) {
	list := todotxt.ListOf()
	list.AddHook(hook.NewRecurrence(defaultTags))

	err := list.Add(
		todotxt.MustBuildItem(todotxt.WithDescription("Hello world rec:+1y")),
	)

	assert.Error(t, err)
}

func Test_ValidationIsTriggeredOnModification(t *testing.T) {
	list := todotxt.ListOf()
	list.AddHook(hook.NewRecurrence(defaultTags))

	err := list.Add(
		todotxt.MustBuildItem(todotxt.WithDescription("Hello world")),
	)
	assert.Nil(t, err)

	err = list.GetLine(1).EditDescription("Hello world rec:5y")
	assert.Error(t, err)
}
