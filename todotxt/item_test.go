package todotxt_test

import (
	"testing"
	"time"

	. "github.com/Fabian-G/todotxt/todotxt"
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
			item := Apply(Create(tc.cDate), WithNowFunc(now))
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
			item:           DummyItem(WithEmptyDescription),
			expectedFormat: "",
		},
		"A task with nothing but a description": {
			item:           EmptyItem(WithDescription("this is a test")),
			expectedFormat: "this is a test",
		},
		"A full dummy task": {
			item:           DummyItem(),
			expectedFormat: "x (F) 2023-04-05 2020-04-05 This is a dummy task",
		},
		"An uncompleted task": {
			item:           DummyItem(Uncompleted),
			expectedFormat: "(F) 2020-04-05 This is a dummy task",
		},
		"A task without priority": {
			item:           DummyItem(WithoutPriority),
			expectedFormat: "x 2023-04-05 2020-04-05 This is a dummy task",
		},
		"A task without any dates": {
			item:           DummyItem(WithoutCreationDate, WithoutCompletionDate),
			expectedFormat: "x (F) This is a dummy task",
		},
		"Description with x in the beginning should start with space": {
			item:           EmptyItem(WithDescription("x test")),
			expectedFormat: " x test",
		},
		"Description with date in the beginning should start with space": {
			item:           EmptyItem(WithDescription("2012-03-04 test")),
			expectedFormat: " 2012-03-04 test",
		},
		"Description with Prio in the beginning should start with space": {
			item:           EmptyItem(WithDescription("(A) test")),
			expectedFormat: " (A) test",
		},
		"Description with x in the beginning, but without space should not start with space": {
			item:           EmptyItem(WithDescription("xTest")),
			expectedFormat: "xTest",
		},
		"Description with date in the beginning, but without space should not start with space": {
			item:           EmptyItem(WithDescription("2012-03-04Test")),
			expectedFormat: "2012-03-04Test",
		},
		"Description with Prio in the beginning, but without space should not start with space": {
			item:           EmptyItem(WithDescription("(A)test")),
			expectedFormat: "(A)test",
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
			item:             DummyItem(WithDescription("A description without projects")),
			expectedProjects: []Project{},
		},
		"A Project defined in the beginning": {
			item:             DummyItem(WithDescription("+projectFoo A description with projects")),
			expectedProjects: []Project{"+projectFoo"},
		},
		"A Project defined in the end": {
			item:             DummyItem(WithDescription("A description with projects +projectFoo")),
			expectedProjects: []Project{"+projectFoo"},
		},
		"A Project defined in middle": {
			item:             DummyItem(WithDescription("A description +projectFoo with projects")),
			expectedProjects: []Project{"+projectFoo"},
		},
		"Multiple Projects defined": {
			item:             DummyItem(WithDescription("+projectFoo A description +projectBar with projects +projectBaz")),
			expectedProjects: []Project{"+projectFoo", "+projectBar", "+projectBaz"},
		},
		"A plus sign within a word": {
			item:             DummyItem(WithDescription("A description with this+is not a project")),
			expectedProjects: []Project{},
		},
		"A plus sign within a project name": {
			item:             DummyItem(WithDescription("A description with +this+is a project")),
			expectedProjects: []Project{"+this+is"},
		},
		"Duplicate Projects only occur once": {
			item:             DummyItem(WithDescription("A description +foo with duplicate +foo projects")),
			expectedProjects: []Project{"+foo"},
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
			item:             DummyItem(WithDescription("A description without contexts")),
			expectedContexts: []Context{},
		},
		"A Context defined in the beginning": {
			item:             DummyItem(WithDescription("@contextFoo A description with contexts")),
			expectedContexts: []Context{"@contextFoo"},
		},
		"A Context defined in the end": {
			item:             DummyItem(WithDescription("A description with Contexts @contextFoo")),
			expectedContexts: []Context{"@contextFoo"},
		},
		"A Context defined in middle": {
			item:             DummyItem(WithDescription("A description @contextFoo with contexts")),
			expectedContexts: []Context{"@contextFoo"},
		},
		"Multiple Contexts defined": {
			item:             DummyItem(WithDescription("@contextFoo A description @contextBar with Contexts @contextBaz")),
			expectedContexts: []Context{"@contextFoo", "@contextBar", "@contextBaz"},
		},
		"An at sign within a word": {
			item:             DummyItem(WithDescription("A description with this@is not a context")),
			expectedContexts: []Context{},
		},
		"An at sign within a context name": {
			item:             DummyItem(WithDescription("A description with @this@is a context")),
			expectedContexts: []Context{"@this@is"},
		},
		"Duplicate Contexts only occur once": {
			item:             DummyItem(WithDescription("A description with @foo duplicate @foo contexts")),
			expectedContexts: []Context{"@foo"},
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
			item:         DummyItem(WithDescription("This is a description")),
			expectedTags: Tags{},
		},
		"A tag at the beginning": {
			item:         DummyItem(WithDescription("foo:bar This is a description")),
			expectedTags: Tags{"foo": {"bar"}},
		},
		"A tag at the end": {
			item:         DummyItem(WithDescription("This is a description foo:bar")),
			expectedTags: Tags{"foo": {"bar"}},
		},
		"A tag in the middle": {
			item:         DummyItem(WithDescription("This is a foo:bar description")),
			expectedTags: Tags{"foo": {"bar"}},
		},
		"Multiple tags with different names": {
			item:         DummyItem(WithDescription("foo:bar This is a bar:baz description baz:foo")),
			expectedTags: Tags{"foo": {"bar"}, "bar": {"baz"}, "baz": {"foo"}},
		},
		"Multiple tags with the same name": {
			item:         DummyItem(WithDescription("foo:foo This is a foo:bar description foo:baz")),
			expectedTags: Tags{"foo": {"foo", "bar", "baz"}},
		},
		"Colon within a context": {
			item:         DummyItem(WithDescription("This is a @foo:bar description")),
			expectedTags: Tags{},
		},
		"Colon within a project": {
			item:         DummyItem(WithDescription("This is a +foo:bar description")),
			expectedTags: Tags{},
		},
		"The key may be empty": {
			item:         DummyItem(WithDescription(":foo This :baz is a description :bar")),
			expectedTags: Tags{"": {"foo", "baz", "bar"}},
		},
		"The value may not be empty": {
			item:         DummyItem(WithDescription("baz: This is a foo: description bar:")),
			expectedTags: Tags{},
		},
		"Tag value can contain colons and other special characters": {
			item:         DummyItem(WithDescription("This is a foo:baz@bar:fo+o description")),
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
	item := &Item{}
	item.Complete()
	item = Apply(item, WithoutCreationDate)

	assert.Panics(t, func() {
		_ = item.String()
	})
}
