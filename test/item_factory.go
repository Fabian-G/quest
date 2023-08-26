package test

import (
	"time"

	"github.com/Fabian-G/todotxt/todotxt"
)

func Apply(item *todotxt.Item, modifier ...func(*todotxt.Item) *todotxt.Item) *todotxt.Item {
	copy := *item
	item = &copy
	for _, m := range modifier {
		item = m(item)
	}
	return item
}

func EmptyItem(modifier ...func(*todotxt.Item) *todotxt.Item) *todotxt.Item {
	return Apply(&todotxt.Item{}, modifier...)
}

func DummyItem(modifier ...func(*todotxt.Item) *todotxt.Item) *todotxt.Item {
	creationTime := time.Date(2020, 4, 5, 6, 7, 8, 9, time.UTC)
	completionTime := time.Date(2023, 4, 5, 6, 7, 8, 9, time.UTC)
	item := MustBuild(todotxt.WithMeta(true, todotxt.PrioNone, completionTime, creationTime), todotxt.WithDescription("This is a dummy task"))
	return Apply(item, modifier...)
}

func WithEmptyDescription(item *todotxt.Item) *todotxt.Item {
	item.EditDescription("")
	return item
}

func Uncompleted(item *todotxt.Item) *todotxt.Item {
	item.MarkUndone()
	return item
}

func WithoutPriority(item *todotxt.Item) *todotxt.Item {
	item.PrioritizeAs(todotxt.PrioNone)
	return item
}

func WithoutCompletionDate(item *todotxt.Item) *todotxt.Item {
	return MustBuild(todotxt.CopyOf(item), todotxt.WithCompletionDate(nil))
}

func WithoutCreationDate(item *todotxt.Item) *todotxt.Item {
	return MustBuild(todotxt.CopyOf(item), todotxt.WithCreationDate(nil))
}

func MustBuild(m ...todotxt.BuildFunc) *todotxt.Item {
	i, err := todotxt.Build(m...)
	if err != nil {
		panic(err)
	}
	return i
}
