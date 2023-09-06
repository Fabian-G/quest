package todotxt_test

import (
	"testing"
	"time"

	"github.com/Fabian-G/quest/todotxt"
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
		baseItem               *todotxt.Item
		expectedCDate          time.Time
		expectedCompletionDate time.Time
	}{
		"Zero CreationDate should be updated to CompletionDate": {
			baseItem:               new(todotxt.Item),
			expectedCDate:          time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC),
			expectedCompletionDate: time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC),
		},
		"CreationDate before CompletionDate should be left untouched": {
			baseItem:               todotxt.MustBuildItem(todotxt.WithCreationDate(time.Date(2023, 8, 20, 12, 3, 4, 1, time.UTC))),
			expectedCDate:          time.Date(2023, 8, 20, 0, 0, 0, 0, time.UTC),
			expectedCompletionDate: time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC),
		},
		"CreationDate after CompletionDate should be updated to completionDate": {
			baseItem:               todotxt.MustBuildItem(todotxt.WithCreationDate(time.Date(2023, 8, 23, 12, 3, 4, 1, time.UTC))),
			expectedCDate:          time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC),
			expectedCompletionDate: time.Date(2023, 8, 22, 0, 0, 0, 0, time.UTC),
		},
	}

	now := nowFuncForDay("2023-08-22")
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			item := todotxt.MustBuildItem(todotxt.CopyOf(tc.baseItem), todotxt.WithNowFunc(now), todotxt.WithDescription("dummy"))
			item.Complete()
			assert.True(t, item.Done())
			assert.Equal(t, tc.expectedCompletionDate, *item.CompletionDate())
			assert.Equal(t, tc.expectedCDate, *item.CreationDate())
		})
	}
}

func Test_ProjectExtraction(t *testing.T) {
	testCases := map[string]struct {
		item             *todotxt.Item
		expectedProjects []todotxt.Project
	}{
		"No Projects defined": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description without projects")),
			expectedProjects: []todotxt.Project{},
		},
		"A Project defined in the beginning": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("+projectFoo A description with projects")),
			expectedProjects: []todotxt.Project{"projectFoo"},
		},
		"A Project defined in the end": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description with projects +projectFoo")),
			expectedProjects: []todotxt.Project{"projectFoo"},
		},
		"A Project defined in middle": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description +projectFoo with projects")),
			expectedProjects: []todotxt.Project{"projectFoo"},
		},
		"Multiple Projects defined": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("+projectFoo A description +projectBar with projects +projectBaz")),
			expectedProjects: []todotxt.Project{"projectFoo", "projectBar", "projectBaz"},
		},
		"A plus sign within a word": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description with this+is not a project")),
			expectedProjects: []todotxt.Project{},
		},
		"A plus sign within a project name": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description with +this+is a project")),
			expectedProjects: []todotxt.Project{"this+is"},
		},
		"Duplicate Projects only occur once": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description +foo with duplicate +foo projects")),
			expectedProjects: []todotxt.Project{"foo"},
		},
		"Multiple contexts in a row can occur": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description with +foo +project")),
			expectedProjects: []todotxt.Project{"foo", "project"},
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
		item             *todotxt.Item
		expectedContexts []todotxt.Context
	}{
		"No Contexts defined": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description without contexts")),
			expectedContexts: []todotxt.Context{},
		},
		"A Context defined in the beginning": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("@contextFoo A description with contexts")),
			expectedContexts: []todotxt.Context{"contextFoo"},
		},
		"A Context defined in the end": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description with Contexts @contextFoo")),
			expectedContexts: []todotxt.Context{"contextFoo"},
		},
		"A Context defined in middle": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description @contextFoo with contexts")),
			expectedContexts: []todotxt.Context{"contextFoo"},
		},
		"Multiple Contexts defined": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("@contextFoo A description @contextBar with Contexts @contextBaz")),
			expectedContexts: []todotxt.Context{"contextFoo", "contextBar", "contextBaz"},
		},
		"An at sign within a word": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description with this@is not a context")),
			expectedContexts: []todotxt.Context{},
		},
		"An at sign within a context name": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description with @this@is a context")),
			expectedContexts: []todotxt.Context{"this@is"},
		},
		"Duplicate Contexts only occur once": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description with @foo duplicate @foo contexts")),
			expectedContexts: []todotxt.Context{"foo"},
		},
		"Multiple contexts in a row can occur": {
			item:             todotxt.MustBuildItem(todotxt.WithDescription("A description with @foo @context")),
			expectedContexts: []todotxt.Context{"foo", "context"},
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
		item         *todotxt.Item
		expectedTags todotxt.Tags
	}{
		"No tags defined": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("This is a description")),
			expectedTags: todotxt.Tags{},
		},
		"A tag at the beginning": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("foo:bar This is a description")),
			expectedTags: todotxt.Tags{"foo": {"bar"}},
		},
		"A tag at the end": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("This is a description foo:bar")),
			expectedTags: todotxt.Tags{"foo": {"bar"}},
		},
		"A tag in the middle": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("This is a foo:bar description")),
			expectedTags: todotxt.Tags{"foo": {"bar"}},
		},
		"Multiple tags with different names": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("foo:bar This is a bar:baz description baz:foo")),
			expectedTags: todotxt.Tags{"foo": {"bar"}, "bar": {"baz"}, "baz": {"foo"}},
		},
		"Multiple tags with the same name": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("foo:foo This is a foo:bar description foo:baz")),
			expectedTags: todotxt.Tags{"foo": {"foo", "bar", "baz"}},
		},
		"Colon within a context": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("This is a @foo:bar description")),
			expectedTags: todotxt.Tags{},
		},
		"Colon within a project": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("This is a +foo:bar description")),
			expectedTags: todotxt.Tags{},
		},
		"The key may be empty": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription(":foo This :baz is a description :bar")),
			expectedTags: todotxt.Tags{"": {"foo", "baz", "bar"}},
		},
		"The value may not be empty": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("baz: This is a foo: description bar:")),
			expectedTags: todotxt.Tags{},
		},
		"Tag value can contain colons and other special characters": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("This is a foo:baz@bar:fo+o description")),
			expectedTags: todotxt.Tags{"foo": {"baz@bar:fo+o"}},
		},
		"Tag value can contain a date string": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("This is a description with due:2022-02-02")),
			expectedTags: todotxt.Tags{"due": {"2022-02-02"}},
		},
		"Description can contain more tags in a row": {
			item:         todotxt.MustBuildItem(todotxt.WithDescription("This is a description with rec:+5d due:2022-02-02")),
			expectedTags: todotxt.Tags{"due": {"2022-02-02"}, "rec": {"+5d"}},
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

func Test_SetTag(t *testing.T) {
	item := todotxt.MustBuildItem(todotxt.WithDescription("This is rec:+5y a description with tags rec:5y rec:+6y"))

	item.SetTag("rec", "+1d")

	assert.Equal(t, "This is rec:+1d a description with tags rec:+1d rec:+1d", item.Description())
}

func Test_SetTagForNewTag(t *testing.T) {
	item := todotxt.MustBuildItem(todotxt.WithDescription("This is rec:+5y a description with tags rec:5y"))

	item.SetTag("new", "+1d")

	assert.Equal(t, "This is rec:+5y a description with tags rec:5y new:+1d", item.Description())
}
