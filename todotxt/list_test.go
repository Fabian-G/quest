package todotxt_test

import (
	"testing"

	"github.com/Fabian-G/quest/todotxt"
	"github.com/stretchr/testify/assert"
)

func Test_HooksGetCalledOnModification(t *testing.T) {
	testCases := map[string]struct {
		initial         *todotxt.List
		op              func(l *todotxt.List)
		previousMatcher func(*testing.T, *todotxt.Item)
		currentMatcher  func(*testing.T, *todotxt.Item)
	}{
		"Hook gets called on add": {
			initial: &todotxt.List{},
			op: func(l *todotxt.List) {
				l.Add(todotxt.MustBuildItem(todotxt.WithDescription("Hello World")))
			},
			previousMatcher: func(t *testing.T, previous *todotxt.Item) {
				assert.Nil(t, previous)
			},
			currentMatcher: func(t *testing.T, current *todotxt.Item) {
				assert.Equal(t, "Hello World", current.Description())
			},
		},
		"Hook gets called on delete": {
			initial: todotxt.ListOf(
				todotxt.MustBuildItem(todotxt.WithDescription("Hello World")),
			),
			op: func(l *todotxt.List) {
				l.Remove(1)
			},
			previousMatcher: func(t *testing.T, previous *todotxt.Item) {
				assert.Equal(t, "Hello World", previous.Description())
			},
			currentMatcher: func(t *testing.T, current *todotxt.Item) {
				assert.Nil(t, current)
			},
		},
		"Hook gets called on complete": {
			initial: todotxt.ListOf(
				todotxt.MustBuildItem(todotxt.WithDescription("Hello World")),
			),
			op: func(l *todotxt.List) {
				l.Get(1).Complete()
			},
			previousMatcher: func(t *testing.T, i *todotxt.Item) {
				assert.Equal(t, "Hello World", i.Description())
				assert.False(t, i.Done())
			},
			currentMatcher: func(t *testing.T, i *todotxt.Item) {
				assert.True(t, i.Done())
			},
		},
		"Hook gets called on mark undone": {
			initial: todotxt.ListOf(
				todotxt.MustBuildItem(todotxt.WithDescription("Hello World"), todotxt.WithDone(true)),
			),
			op: func(l *todotxt.List) {
				l.Get(1).MarkUndone()
			},
			previousMatcher: func(t *testing.T, i *todotxt.Item) {
				assert.Equal(t, "Hello World", i.Description())
				assert.True(t, i.Done())
			},
			currentMatcher: func(t *testing.T, i *todotxt.Item) {
				assert.False(t, i.Done())
			},
		},
		"Hook gets called on description change": {
			initial: todotxt.ListOf(
				todotxt.MustBuildItem(todotxt.WithDescription("Hello World")),
			),
			op: func(l *todotxt.List) {
				l.Get(1).EditDescription("This is a description change")
			},
			previousMatcher: func(t *testing.T, i *todotxt.Item) {
				assert.Equal(t, "Hello World", i.Description())
			},
			currentMatcher: func(t *testing.T, i *todotxt.Item) {
				assert.Equal(t, "This is a description change", i.Description())
			},
		},
		"Hook gets called on priority change": {
			initial: todotxt.ListOf(
				todotxt.MustBuildItem(todotxt.WithDescription("Hello World")),
			),
			op: func(l *todotxt.List) {
				l.Get(1).PrioritizeAs(todotxt.PrioB)
			},
			previousMatcher: func(t *testing.T, i *todotxt.Item) {
				assert.Equal(t, todotxt.PrioNone, i.Priority())
			},
			currentMatcher: func(t *testing.T, i *todotxt.Item) {
				assert.Equal(t, todotxt.PrioB, i.Priority())
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var hookCalled bool
			tc.initial.AddHook(
				todotxt.HookFunc(func(list *todotxt.List, event todotxt.ModEvent) error {
					hookCalled = true
					tc.previousMatcher(t, event.Previous)
					tc.currentMatcher(t, event.Current)
					return nil
				}),
			)
			tc.op(tc.initial)
			assert.True(t, hookCalled)
		})
	}
}

func Test_ModificationsMadeInTheHookGetThroughToTheList(t *testing.T) {
	list := todotxt.ListOf(todotxt.MustBuildItem(todotxt.WithDescription("Hello World")))
	undoer := todotxt.HookFunc(func(list *todotxt.List, event todotxt.ModEvent) error {
		// We don't like any changes, so we undo them all, but leave a note
		*event.Current = *event.Previous
		event.Current.EditDescription(event.Current.Description() + " That change sucked")
		return nil
	})
	list.AddHook(undoer)

	list.Get(1).Complete()
	list.Get(1).EditDescription("Foo")
	list.Get(1).PrioritizeAs(todotxt.PrioA)

	assert.Equal(t, "Hello World That change sucked That change sucked That change sucked", list.Get(1).Description())
	assert.False(t, list.Get(1).Done())
}

func TestList_Remove(t *testing.T) {
	list := todotxt.ListOf(todotxt.MustBuildItem(todotxt.WithDescription("Hello World")))

	err := list.Remove(1)

	assert.Nil(t, err)
	assert.Equal(t, 0, list.Len())
}

func TestList_Add(t *testing.T) {
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("A")),
		todotxt.MustBuildItem(todotxt.WithDescription("C")),
	)

	newItem := todotxt.MustBuildItem(todotxt.WithDescription("B"))
	err := list.Add(newItem)

	assert.Nil(t, err)
	assert.Equal(t, 3, list.IndexOf(newItem))
}

func Test_SnapshotAndRestore(t *testing.T) {
	list := todotxt.ListOf(
		todotxt.MustBuildItem(todotxt.WithDescription("A")),
		todotxt.MustBuildItem(todotxt.WithDescription("B")),
	)
	pointerA := list.Get(1)
	list.Snapshot()

	err := list.Get(1).EditDescription("Hello World")
	assert.Nil(t, err)
	err = list.Add(todotxt.MustBuildItem(todotxt.WithDescription("123")))
	assert.Nil(t, err)
	err = list.Add(todotxt.MustBuildItem(todotxt.WithDescription("12345")))
	assert.Nil(t, err)
	err = list.Remove(2)
	assert.Nil(t, err)

	list.Reset()

	assert.Equal(t, 2, list.Len())
	assert.Equal(t, "A", pointerA.Description())
	assert.Equal(t, "B", list.Get(2).Description())

}
