package todotxt

import (
	"errors"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/repeale/fp-go"
)

// Used to determine whether or not the completionDate/creationDate was initialized
var zeroTime = time.Time{}

// Tests whether or not a description needs a leading space when formatted to avoid being ambiguous.
// For example when serializing the description "x test" it would be deserialized to {done: true, desc: test} unless
// we add an additional space when serializing: " x test"
var leadingSpaceNeeded = regexp.MustCompile("^x |^[0-9]{4}-[0-9]{2}-[0-9]{2} |^\\([A-Z]\\) ")

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

func (i *Item) Description() string {
	return i.description
}

func (i *Item) Projects() []Project {
	matches := projectRegex.FindAllString(i.description, -1)
	toProject := fp.Map(func(s string) Project { return Project(strings.TrimSpace(s)) })
	sort := func(in []Project) []Project {
		slices.Sort(in)
		return in
	}
	uniq := slices.Compact[[]Project]
	return fp.Pipe3(toProject, sort, uniq)(matches)
}

func (i *Item) Contexts() []Context {
	matches := contextRegex.FindAllString(i.description, -1)
	toContext := fp.Map(func(s string) Context { return Context(strings.TrimSpace(s)) })
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
	i.completionDate = truncateToDate(i.now())
	if i.creationDate == zeroTime || i.creationDate.After(i.completionDate) {
		i.creationDate = i.completionDate
	}
}

func (i *Item) MarkUndone() {
	i.done = false
	i.completionDate = time.Time{}
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

// String formats the todo item according to the todotxt spec.
// This method is not (only) for convenient printing, but is also used for persistence.
// Therefore this *must* adhere to the todotxt spec.
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
	if builder.Len() == 0 && leadingSpaceNeeded.MatchString(i.description) {
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
