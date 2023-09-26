package hook_test

import (
	"testing"
	"time"

	"github.com/Fabian-G/quest/hook"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

func Test_IntExpansion(t *testing.T) {
	testCases := map[string]struct {
		otherTasksInList  []*todotxt.Item
		description       string
		expectedExpansion string
	}{
		"int without expansion is left unchanged": {
			description:       "hello i:-4 world",
			expectedExpansion: "hello i:-4 world",
		},
		"expand maximum": {
			otherTasksInList: []*todotxt.Item{
				todotxt.MustBuildItem(todotxt.WithDescription("large i:12 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small i:-4 int")),
			},
			description:       "hello i:max world",
			expectedExpansion: "hello i:12 world",
		},
		"expand minimum": {
			otherTasksInList: []*todotxt.Item{
				todotxt.MustBuildItem(todotxt.WithDescription("large i:12 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small i:-4 int")),
			},
			description:       "hello i:min world",
			expectedExpansion: "hello i:-4 world",
		},
		"expand project maximum": {
			otherTasksInList: []*todotxt.Item{
				todotxt.MustBuildItem(todotxt.WithDescription("large +foo i:12 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("large of other i:24 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small +foo i:-4 int")),
			},
			description:       "hello +foo i:pmax world",
			expectedExpansion: "hello +foo i:12 world",
		},
		"expand project minimum": {
			otherTasksInList: []*todotxt.Item{
				todotxt.MustBuildItem(todotxt.WithDescription("large +foo i:12 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small +foo i:4 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small of other i:-24 int")),
			},
			description:       "hello +foo i:pmin world",
			expectedExpansion: "hello +foo i:4 world",
		},
		"expand project minimum without project is always 0": {
			otherTasksInList: []*todotxt.Item{
				todotxt.MustBuildItem(todotxt.WithDescription("large +foo i:12 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small +foo i:-4 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small of other i:-24 int")),
			},
			description:       "hello i:pmin world",
			expectedExpansion: "hello i:0 world",
		},
		"expand maximum with positive offset": {
			otherTasksInList: []*todotxt.Item{
				todotxt.MustBuildItem(todotxt.WithDescription("large +foo i:12 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small +foo i:-4 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small of other i:-24 int")),
			},
			description:       "hello i:max+3 world",
			expectedExpansion: "hello i:15 world",
		},
		"expand maximum with negative offset": {
			otherTasksInList: []*todotxt.Item{
				todotxt.MustBuildItem(todotxt.WithDescription("large +foo i:12 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small +foo i:-4 int")),
				todotxt.MustBuildItem(todotxt.WithDescription("small of other i:-24 int")),
			},
			description:       "hello i:max-3 world",
			expectedExpansion: "hello i:9 world",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			item := todotxt.MustBuildItem(
				todotxt.WithDescription("Hello World"),
			)
			list := todotxt.ListOf(append(tc.otherTasksInList, item)...)
			list.AddHook(hook.NewTagExpansion(false, map[string]qselect.DType{
				"i": qselect.QInt,
			}))

			err := item.EditDescription(tc.description)

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedExpansion, item.Description())
		})
	}
}

func Test_DateExpansion(t *testing.T) {
	today, _ := time.Parse(time.DateOnly, "2022-02-02")
	testCases := map[string]struct {
		otherTasksInList  []*todotxt.Item
		description       string
		expectedExpansion string
	}{
		"date without expansion is left unchanged": {
			description:       "hello due:2023-01-01 world",
			expectedExpansion: "hello due:2023-01-01 world",
		},
		"today is inserted": {
			description:       "hello due:today world",
			expectedExpansion: "hello due:2022-02-02 world",
		},
		"positive duration offset": {
			description:       "hello due:+1d world",
			expectedExpansion: "hello due:2022-02-03 world",
		},
		"negative duration offset": {
			description:       "hello due:-1d world",
			expectedExpansion: "hello due:2022-02-01 world",
		},
		"positive duration offset with base": {
			description:       "hello due:today+1d world",
			expectedExpansion: "hello due:2022-02-03 world",
		},
		"negative duration offset with base": {
			description:       "hello due:today-1d world",
			expectedExpansion: "hello due:2022-02-01 world",
		},
		"unknown tags are allowed": {
			description:       "an unknown:tag should work",
			expectedExpansion: "an unknown:tag should work",
		},
		"monday": {
			description:       "due:monday",
			expectedExpansion: "due:2022-02-07",
		},
		"tuesday": {
			description:       "due:tuesday",
			expectedExpansion: "due:2022-02-08",
		},
		"wednesday": {
			description:       "due:wednesday",
			expectedExpansion: "due:2022-02-02",
		},
		"thursday": {
			description:       "due:thursday",
			expectedExpansion: "due:2022-02-03",
		},
		"friday": {
			description:       "due:friday",
			expectedExpansion: "due:2022-02-04",
		},
		"saturday": {
			description:       "due:saturday",
			expectedExpansion: "due:2022-02-05",
		},
		"sunday": {
			description:       "due:sunday",
			expectedExpansion: "due:2022-02-06",
		},
		"tomorrow": {
			description:       "due:tomorrow",
			expectedExpansion: "due:2022-02-03",
		},
		"yesterday": {
			description:       "due:yesterday",
			expectedExpansion: "due:2022-02-01",
		},
		"case insensitive": {
			description:       "due:MoNdAy",
			expectedExpansion: "due:2022-02-07",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			item := todotxt.MustBuildItem(
				todotxt.WithDescription("Hello World"),
			)
			list := todotxt.ListOf(append(tc.otherTasksInList, item)...)
			list.AddHook(hook.NewTagExpansionWithNowFunc(true, map[string]qselect.DType{
				"due": qselect.QDate,
			}, func() time.Time {
				return today
			}))

			err := item.EditDescription(tc.description)

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedExpansion, item.Description())
		})
	}
}

func Test_TagExpansionsIgnoresRemovalEvents(t *testing.T) {
	item := todotxt.MustBuildItem(
		todotxt.WithDescription("Hello World"),
	)
	list := todotxt.ListOf(item)
	list.AddHook(hook.NewTagExpansionWithNowFunc(false, map[string]qselect.DType{
		"due": qselect.QDate,
	}, func() time.Time {
		return time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)
	}))

	err := list.Remove(list.IndexOf(item))

	assert.Nil(t, err)
	assert.Equal(t, 0, list.Len())
}

func Test_TagExpansionsError(t *testing.T) {
	testCases := map[string]struct {
		description string
	}{
		"unknown tags": {
			description: "hello unknown:tag abc",
		},
		"wrong expansion": {
			description: "a wrong int someint:expansion",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			item := todotxt.MustBuildItem(
				todotxt.WithDescription("Hello World"),
			)
			list := todotxt.ListOf(item)
			list.AddHook(hook.NewTagExpansionWithNowFunc(false, map[string]qselect.DType{
				"due":     qselect.QDate,
				"someInt": qselect.QInt,
			}, func() time.Time {
				return time.Date(2022, 2, 2, 0, 0, 0, 0, time.UTC)
			}))

			err := item.EditDescription(tc.description)

			assert.Error(t, err)
		})
	}

}
