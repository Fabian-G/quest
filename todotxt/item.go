package todotxt

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/repeale/fp-go"
)

type List struct {
	version   time.Time
	todoItems []Item
}

var ErrCreationDateUnset = errors.New("completion date can not be set while creation date is not")
var ErrCompleteBeforeCreation = errors.New("completion date can not be before creation date")
var ErrCompletionDateWhileUndone = errors.New("completion date can not be set on undone task")
var ErrEmptyDescription = errors.New("description must not be empty")
var ErrNoPrioWhenDone = errors.New("done tasks must not have a priority")

type Item struct {
	nowFunc        func() time.Time
	done           bool
	prio           Priority
	completionDate *time.Time
	creationDate   *time.Time
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

func (i *Item) CompletionDate() *time.Time {
	return i.completionDate
}

func (i *Item) CreationDate() *time.Time {
	return i.creationDate
}

func (i *Item) Description() string {
	return i.description
}

func (i *Item) Projects() []Project {
	matches := projectRegex.FindAllString(i.description, -1)
	toProject := fp.Map(func(s string) Project { return Project(strings.TrimSpace(s)[1:]) })
	sort := func(in []Project) []Project {
		slices.Sort(in)
		return in
	}
	uniq := slices.Compact[[]Project]
	return fp.Pipe3(toProject, sort, uniq)(matches)
}

func (i *Item) Contexts() []Context {
	matches := contextRegex.FindAllString(i.description, -1)
	toContext := fp.Map(func(s string) Context { return Context(strings.TrimSpace(s)[1:]) })
	sort := func(in []Context) []Context {
		slices.Sort(in)
		return in
	}
	uniq := slices.Compact[[]Context]
	return fp.Pipe3(toContext, sort, uniq)(matches)
}

func (i *Item) Tags() Tags {
	type tag struct {
		key   string
		value string
	}
	matches := tagRegex.FindAllString(i.description, -1)
	split := fp.Map(func(match string) tag {
		tagSepIndex := strings.Index(match, ":")
		return tag{
			key:   strings.TrimSpace(match[:tagSepIndex]),
			value: strings.TrimSpace(match[tagSepIndex+1:]),
		}
	})
	toMap := fp.Reduce(func(tags Tags, t tag) Tags {
		tags[t.key] = append(tags[t.key], t.value)
		return tags
	}, Tags(make(map[string][]string)))
	return fp.Pipe2(split, toMap)(matches)
}

func (i *Item) Complete() {
	i.done = true
	i.prio = PrioNone
	i.completionDate = truncateToDate(i.now())
	if i.creationDate == nil || i.creationDate.After(*i.completionDate) {
		i.creationDate = i.completionDate
	}
}

func (i *Item) MarkUndone() {
	i.done = false
	i.completionDate = nil
}

func (i *Item) PrioritizeAs(prio Priority) {
	i.done = false
	i.prio = prio
}

func (i *Item) EditDescription(desc string) error {
	if len(strings.TrimSpace(desc)) == 0 {
		return ErrEmptyDescription
	}
	i.description = desc
	return nil
}

func (i *Item) String() string {
	if err := i.valid(); err != nil {
		return fmt.Sprintf("%#v", i)
	}
	return DefaultFormatter.Format(i)
}

// This method is unexported, because the API is designed in a way that should make
// it impossible for the user t create invalid tasks
func (i *Item) valid() error {
	if len(strings.TrimSpace(i.description)) == 0 {
		return ErrEmptyDescription
	}
	if i.completionDate != nil && i.creationDate == nil {
		return ErrCreationDateUnset
	}
	if i.completionDate != nil && i.creationDate != nil && i.completionDate.Before(*i.CreationDate()) {
		return ErrCompleteBeforeCreation
	}
	if !i.done && i.completionDate != nil {
		return ErrCompletionDateWhileUndone
	}
	if i.done && i.prio != PrioNone {
		return ErrNoPrioWhenDone
	}
	return nil
}

func truncateToDate(t time.Time) *time.Time {
	truncatedDate := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return &truncatedDate
}

func (i *Item) now() time.Time {
	if i.nowFunc != nil {
		return i.nowFunc()
	}
	return time.Now()
}
