package todotxt

import (
	"time"
)

type BuildFunc func(*Item) *Item

func Build(modifier ...BuildFunc) (*Item, error) {
	item := &Item{}
	for _, m := range modifier {
		item = m(item)
	}
	return item, item.valid()
}

func MustBuild(m ...BuildFunc) *Item {
	i, err := Build(m...)
	if err != nil {
		panic(err)
	}
	return i
}

func WithDescription(desc string) BuildFunc {
	return func(i *Item) *Item {
		i.description = desc
		return i
	}
}

func WithMeta(done bool, prio Priority, completionDate time.Time, creationDate time.Time) BuildFunc {
	return func(i *Item) *Item {
		i.done = done
		i.completionDate = truncateToDate(completionDate)
		i.creationDate = truncateToDate(creationDate)
		i.prio = prio
		return i
	}
}

func WithDone(done bool) BuildFunc {
	return func(i *Item) *Item {
		i.done = done
		return i
	}
}

func WithCreationDate(date *time.Time) BuildFunc {
	return func(i *Item) *Item {
		if date == nil {
			i.creationDate = nil
			return i
		}
		i.creationDate = truncateToDate(*date)
		return i
	}
}

func WithCompletionDate(date *time.Time) BuildFunc {
	return func(i *Item) *Item {
		if date == nil {
			i.completionDate = nil
			return i
		}
		i.completionDate = truncateToDate(*date)
		return i
	}
}

func WithPriority(prio Priority) BuildFunc {
	return func(i *Item) *Item {
		i.prio = prio
		return i
	}
}

func WithNowFunc(now func() time.Time) BuildFunc {
	return func(i *Item) *Item {
		i.nowFunc = now
		return i
	}
}

func CopyOf(item *Item) BuildFunc {
	return func(i *Item) *Item {
		copy := *item
		return &copy
	}
}
