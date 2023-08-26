package todotxt

import (
	"slices"
	"strings"
	"time"

	"github.com/repeale/fp-go"
)

type List struct {
	version   time.Time
	todoItems []Item
}

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
	i.prio = prio
}

func (i *Item) EditDescription(desc string) {
	i.description = desc
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
