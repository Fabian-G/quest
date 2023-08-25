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

func Test_ProjectExtraction(t *testing.T) {
	testCases := map[string]struct {
		item             *Item
		expectedProjects []Project
	}{
		"No Projects defined": {
			item:             dummyTask(withDescription("A description without projects")),
			expectedProjects: []Project{},
		},
		"A Project defined in the beginning": {
			item:             dummyTask(withDescription("+projectFoo A description with projects")),
			expectedProjects: []Project{"+projectFoo"},
		},
		"A Project defined in the end": {
			item:             dummyTask(withDescription("A description with projects +projectFoo")),
			expectedProjects: []Project{"+projectFoo"},
		},
		"A Project defined in middle": {
			item:             dummyTask(withDescription("A description +projectFoo with projects")),
			expectedProjects: []Project{"+projectFoo"},
		},
		"Multiple Projects defined": {
			item:             dummyTask(withDescription("+projectFoo A description +projectBar with projects +projectBaz")),
			expectedProjects: []Project{"+projectFoo", "+projectBar", "+projectBaz"},
		},
		"A plus sign within a word": {
			item:             dummyTask(withDescription("A description with this+is not a project")),
			expectedProjects: []Project{},
		},
		"A plus sign within a project name": {
			item:             dummyTask(withDescription("A description with +this+is a project")),
			expectedProjects: []Project{"+this+is"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.expectedProjects, tc.item.Projects())
		})
	}
}

func Test_ContextExtraction(t *testing.T) {
	testCases := map[string]struct {
		item             *Item
		expectedContexts []Context
	}{
		"No Contexts defined": {
			item:             dummyTask(withDescription("A description without contexts")),
			expectedContexts: []Context{},
		},
		"A Context defined in the beginning": {
			item:             dummyTask(withDescription("@contextFoo A description with contexts")),
			expectedContexts: []Context{"@contextFoo"},
		},
		"A Context defined in the end": {
			item:             dummyTask(withDescription("A description with Contexts @contextFoo")),
			expectedContexts: []Context{"@contextFoo"},
		},
		"A Context defined in middle": {
			item:             dummyTask(withDescription("A description @contextFoo with contexts")),
			expectedContexts: []Context{"@contextFoo"},
		},
		"Multiple Contexts defined": {
			item:             dummyTask(withDescription("@contextFoo A description @contextBar with Contexts @contextBaz")),
			expectedContexts: []Context{"@contextFoo", "@contextBar", "@contextBaz"},
		},
		"An at sign within a word": {
			item:             dummyTask(withDescription("A description with this@is not a context")),
			expectedContexts: []Context{},
		},
		"An at sign within a context name": {
			item:             dummyTask(withDescription("A description with @this@is a context")),
			expectedContexts: []Context{"@this@is"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.expectedContexts, tc.item.Contexts())
		})
	}
}

func Test_TagsExtraction(t *testing.T) {
	testCases := map[string]struct {
		item         *Item
		expectedTags Tags
	}{
		"No tags defined": {
			item:         dummyTask(withDescription("This is a description")),
			expectedTags: Tags{},
		},
		"A tag at the beginning": {
			item:         dummyTask(withDescription("foo:bar This is a description")),
			expectedTags: Tags{"foo": {"bar"}},
		},
		"A tag at the end": {
			item:         dummyTask(withDescription("This is a description foo:bar")),
			expectedTags: Tags{"foo": {"bar"}},
		},
		"A tag in the middle": {
			item:         dummyTask(withDescription("This is a foo:bar description")),
			expectedTags: Tags{"foo": {"bar"}},
		},
		"Multiple tags with different names": {
			item:         dummyTask(withDescription("foo:bar This is a bar:baz description baz:foo")),
			expectedTags: Tags{"foo": {"bar"}, "bar": {"baz"}, "baz": {"foo"}},
		},
		"Multiple tags with the same name": {
			item:         dummyTask(withDescription("foo:foo This is a foo:bar description foo:baz")),
			expectedTags: Tags{"foo": {"foo", "bar", "baz"}},
		},
		"Colon within a context": {
			item:         dummyTask(withDescription("This is a @foo:bar description")),
			expectedTags: Tags{},
		},
		"Colon within a project": {
			item:         dummyTask(withDescription("This is a +foo:bar description")),
			expectedTags: Tags{},
		},
		"The key may be empty": {
			item:         dummyTask(withDescription(":foo This :baz is a description :bar")),
			expectedTags: Tags{"": {"foo", "baz", "bar"}},
		},
		"The value may not be empty": {
			item:         dummyTask(withDescription("baz: This is a foo: description bar:")),
			expectedTags: Tags{},
		},
		"Tag value can contain colons and other special characters": {
			item:         dummyTask(withDescription("This is a foo:baz@bar:fo+o description")),
			expectedTags: Tags{"foo": {"baz@bar:fo+o"}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			tags := tc.item.Tags()
			assert.Equal(t, len(tc.expectedTags), len(tags), "Number of found tags differ from expectation")
			for k, v := range tc.expectedTags {
				assert.ElementsMatch(t, v, tags[k], "Values for tag %s did not match", k)
			}
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

func withDescription(desc string) func(i *Item) {
	return func(i *Item) {
		i.EditDescription(desc)
	}
}
