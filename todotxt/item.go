package todotxt

import (
	"errors"
	"strings"
	"time"
)

// Used to determine whether or not the completionDate/creationDate way initialized
var zeroTime = time.Time{}

type List struct {
	version   time.Time
	todoItems []Item
}

type Item struct {
	nowFunc        func() time.Time
	done           bool
	prio           Priority
	completionDate time.Time
	creationDate   time.Time
	description    string
}

// Create creates a new Item with the creationDate set
// This function is only useful when you want to specify the creationDate
// Otherwise the zero value can be used
func Create(creationDate time.Time) *Item {
	item := Item{}
	item.creationDate = truncateToDate(creationDate)
	return &item
}

func (i *Item) Done() bool {
	return i.done
}

func (i *Item) Priority() Priority {
	return i.prio
}

func (i *Item) CompletionDate() time.Time {
	return i.completionDate
}

func (i *Item) CreationDate() time.Time {
	return i.creationDate
}

func (i *Item) Complete() {
	i.done = true
	i.completionDate = truncateToDate(i.now())
	if i.creationDate == zeroTime || i.creationDate.After(i.completionDate) {
		i.creationDate = i.completionDate
	}
}

func (i *Item) MarkUndone() {
	i.done = false
	i.completionDate = time.Time{}
}

func (i *Item) Description() string {
	return i.description
}

func (i *Item) PrioritizeAs(prio Priority) {
	i.prio = prio
}

func (i *Item) EditDescription(desc string) {
	i.description = desc
}

func truncateToDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func (i *Item) String() string {
	if i.completionDate != zeroTime && i.creationDate == zeroTime {
		// In fact this can not really happen
		panic(errors.New("trying to serialize invalid task. CompletionDate set, but CreationDate is not"))
	}
	if i.description == "" {
		return ""
	}
	builder := strings.Builder{}
	if i.Done() {
		builder.WriteString("x")
		builder.WriteString(" ")
	}
	if i.Priority() != PrioNone {
		builder.WriteString(i.prio.String())
		builder.WriteString(" ")
	}
	if i.completionDate != zeroTime {
		builder.WriteString(i.completionDate.Format(time.DateOnly))
		builder.WriteString(" ")
	}
	if i.creationDate != zeroTime {
		builder.WriteString(i.creationDate.Format(time.DateOnly))
		builder.WriteString(" ")
	}
	builder.WriteString(i.description)
	return builder.String()
}

func (i *Item) now() time.Time {
	if i.nowFunc != nil {
		return i.nowFunc()
	}
	return time.Now()
}
