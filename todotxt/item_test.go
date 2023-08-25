package todotxt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func nowFuncForDay(todays string) func() time.Time {
	return func() time.Time {
		now := time.Now()
		today, _ := time.Parse(time.DateOnly, todays)

		return time.Date(today.Year(), today.Month(), today.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
	}
}

func Test_Complete(t *testing.T) {
	testCases := map[string]struct {
		cDate                  time.Time
		expectedCDate          time.Time
		expectedCompletionDate time.Time
	}{
		"Zero CreationDate should be updated to CompletionDate": {
			cDate:                  time.Time{},
			expectedCDate:          time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC),
			expectedCompletionDate: time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC),
		},
		"CreationDate before CompletionDate should be left untouched": {
			cDate:                  time.Date(2023, 8, 20, 12, 3, 4, 1, time.UTC),
			expectedCDate:          time.Date(2023, 8, 20, 0, 0, 0, 0, time.UTC),
			expectedCompletionDate: time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC),
		},
		"CreationDate after CompletionDate should be updated to completionDate": {
			cDate:                  time.Date(2023, 8, 23, 12, 3, 4, 1, time.UTC),
			expectedCDate:          time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC),
			expectedCompletionDate: time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC),
		},
	}

	now := nowFuncForDay("2023-08-22")
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			item := Create(tc.cDate)
			item.nowFunc = now
			item.Complete()
			assert.True(t, item.Done())
			assert.Equal(t, tc.expectedCompletionDate, item.CompletionDate())
			assert.Equal(t, tc.expectedCDate, item.CreationDate())
		})
	}
}

func Test_String(t *testing.T) {
	testCases := map[string]struct {
		item           *Item
		expectedFormat string
	}{
		"Empty task": {
			item:           &Item{},
			expectedFormat: "",
		},
		"Empty description": {
			item:           dummyTask(withEmptyDescription),
			expectedFormat: "",
		},
		"A task with nothing but a description": {
			item:           &Item{description: "this is a test"},
			expectedFormat: "this is a test",
		},
		"A full dummy task": {
			item:           dummyTask(),
			expectedFormat: "x (F) 2023-04-05 2020-04-05 This is a dummy task",
		},
		"An uncompleted task": {
			item:           dummyTask(uncompleted),
			expectedFormat: "(F) 2020-04-05 This is a dummy task",
		},
		"A task without priority": {
			item:           dummyTask(withoutPriority),
			expectedFormat: "x 2023-04-05 2020-04-05 This is a dummy task",
		},
		"A task without any dates": {
			item:           dummyTask(withoutCreationDate, withoutCompletionDate),
			expectedFormat: "x (F) This is a dummy task",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedFormat, tc.item.String())
		})
	}
}

func Test_StringPanicsOnInvalidTask(t *testing.T) {
	item := Item{completionDate: time.Now()}

	assert.Panics(t, func() {
		_ = item.String()
	})
}

func dummyTask(modifier ...func(*Item)) *Item {
	item := &Item{
		nowFunc:        nil,
		done:           true,
		prio:           PrioF,
		completionDate: time.Date(2023, 4, 5, 6, 7, 8, 9, time.UTC),
		creationDate:   time.Date(2020, 4, 5, 6, 7, 8, 9, time.UTC),
		description:    "This is a dummy task",
	}
	for _, m := range modifier {
		m(item)
	}
	return item
}

func withEmptyDescription(item *Item) {
	item.EditDescription("")
}

func uncompleted(item *Item) {
	item.MarkUndone()
}

func withoutPriority(item *Item) {
	item.PrioritizeAs(PrioNone)
}

func withoutCompletionDate(item *Item) {
	item.completionDate = time.Time{}
}

func withoutCreationDate(item *Item) {
	item.creationDate = time.Time{}
}
