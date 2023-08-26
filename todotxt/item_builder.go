package todotxt

import (
	"errors"
	"time"
)

type BuildFunc func(*Item) *Item

var ErrCreationDateUnset = errors.New("completion date can not be set while creation date is not")
var ErrCompleteBeforeCreation = errors.New("completion date can not be before creation date")
var ErrCompletionDateWhileUndone = errors.New("completion date can not be set on undone task")

func Build(modifier ...BuildFunc) (*Item, error) {
	item := Create(time.Now())
	for _, m := range modifier {
		item = m(item)
	}
	return item, validate(item)
}

func validate(item *Item) error {
	if item.completionDate != nil && item.creationDate == nil {
		return ErrCreationDateUnset
	}
	if item.completionDate != nil && item.creationDate != nil && item.completionDate.Before(*item.CreationDate()) {
		return ErrCompleteBeforeCreation
	}
	if !item.done && item.completionDate != nil {
		return ErrCompletionDateWhileUndone
	}
	return nil
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
