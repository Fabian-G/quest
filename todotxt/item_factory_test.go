package todotxt

import "time"

func Apply(item *Item, modifier ...func(*Item)) *Item {
	for _, m := range modifier {
		m(item)
	}
	return item
}

func EmptyItem(modifier ...func(*Item)) *Item {
	return Apply(&Item{}, modifier...)
}

func DummyItem(modifier ...func(*Item)) *Item {
	item := &Item{
		nowFunc:        nil,
		done:           true,
		prio:           PrioF,
		completionDate: time.Date(2023, 4, 5, 6, 7, 8, 9, time.UTC),
		creationDate:   time.Date(2020, 4, 5, 6, 7, 8, 9, time.UTC),
		description:    "This is a dummy task",
	}
	return Apply(item, modifier...)
}

func WithEmptyDescription(item *Item) {
	item.EditDescription("")
}

func Uncompleted(item *Item) {
	item.MarkUndone()
}

func WithoutPriority(item *Item) {
	item.PrioritizeAs(PrioNone)
}

func WithoutCompletionDate(item *Item) {
	item.completionDate = time.Time{}
}

func WithoutCreationDate(item *Item) {
	item.creationDate = time.Time{}
}

func WithDescription(desc string) func(i *Item) {
	return func(i *Item) {
		i.EditDescription(desc)
	}
}

func WithNowFunc(now func() time.Time) func(i *Item) {
	return func(i *Item) {
		i.nowFunc = now
	}
}
