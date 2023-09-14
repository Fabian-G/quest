package todotxt

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

var tagRegex = regexp.MustCompile("(?:^| )[^[:space:]:@+]*:[^[:space:]]+(?: |$)")
var projectRegex = regexp.MustCompile(`(?:^| )\+[^[:space:]]+(?: |$)`)
var contextRegex = regexp.MustCompile("(?:^| )@[^[:space:]]+(?: |$)")

var ErrCreationDateUnset = errors.New("completion date can not be set while creation date is not")
var ErrCompleteBeforeCreation = errors.New("completion date can not be before creation date")
var ErrCompletionDateWhileUndone = errors.New("completion date can not be set on undone task")
var ErrNoPrioWhenDone = errors.New("done tasks must not have a priority")

type Item struct {
	nowFunc        func() time.Time
	done           bool
	prio           Priority
	completionDate *time.Time
	creationDate   *time.Time
	description    string
	emitFunc       func(Event) error
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

func (i *Item) CleanDescription(projects []Project, contexts []Context, tags []string) string {
	desc := i.Description()
	for _, p := range projects {
		matcher := p.Matcher()
		desc = strings.TrimSpace(matcher.ReplaceAllString(desc, " "))
	}
	for _, c := range contexts {
		matcher := c.Matcher()
		desc = strings.TrimSpace(matcher.ReplaceAllString(desc, " "))
	}
	for _, t := range tags {
		matcher := MatcherForTag(t)
		desc = strings.TrimSpace(matcher.ReplaceAllString(desc, " "))
	}
	return desc
}

func (i *Item) Projects() []Project {
	matches := i.findDescMatches(projectRegex)
	projects := make([]Project, 0, len(matches))
	for _, match := range matches {
		projects = append(projects, Project(strings.TrimSpace(match)[1:]))
	}
	slices.Sort(projects)
	projects = slices.Compact(projects)
	return projects
}

func (i *Item) Contexts() []Context {
	matches := i.findDescMatches(contextRegex)
	contexts := make([]Context, 0, len(matches))
	for _, match := range matches {
		contexts = append(contexts, Context(strings.TrimSpace(match)[1:]))
	}
	slices.Sort(contexts)
	contexts = slices.Compact(contexts)
	return contexts
}

func (i *Item) Tags() Tags {
	matches := i.findDescMatches(tagRegex)
	tags := make(Tags)
	for _, match := range matches {
		tagSepIndex := strings.Index(match, ":")
		key := strings.TrimSpace(match[:tagSepIndex])
		value := strings.TrimSpace(match[tagSepIndex+1:])
		tags[key] = append(tags[key], value)
	}
	return tags
}

func (i *Item) findDescMatches(regex *regexp.Regexp) []string {
	matches := make([]string, 0)
	desc := i.description
	for len(desc) > 0 {
		nextMatch := regex.FindStringIndex(desc)
		if nextMatch == nil {
			break
		}
		matches = append(matches, desc[nextMatch[0]:nextMatch[1]])
		desc = desc[nextMatch[1]:]
	}
	return matches
}

func (i *Item) replaceDescMatches(regex *regexp.Regexp, replacement string) string {
	newDescription := strings.TrimSpace(i.description)
	remainingDesc := strings.TrimSpace(i.description)
	for len(remainingDesc) > 0 {
		nextMatch := regex.FindStringIndex(remainingDesc)
		if nextMatch == nil {
			break
		}
		offset := len(newDescription) - len(remainingDesc)
		newDescription = newDescription[:nextMatch[0]+offset] + replacement + newDescription[nextMatch[1]+offset:]
		newDescription = strings.TrimSpace(newDescription)
		remainingDesc = remainingDesc[nextMatch[1]-1:]
	}
	return newDescription
}

func (i *Item) SetTag(key, value string) error {
	matcher := MatcherForTag(key)
	if matcher.MatchString(i.Description()) {
		newDescription := i.replaceDescMatches(matcher, fmt.Sprintf(" %s:%s ", key, value))
		return i.EditDescription(strings.TrimSpace(newDescription))
	} else {
		return i.EditDescription(fmt.Sprintf("%s %s:%s", i.Description(), key, value))
	}
}

func (i *Item) Complete() error {
	return i.modify(func() {
		i.done = true
		i.prio = PrioNone
		i.completionDate = truncateToDate(i.now())
		if i.creationDate == nil || i.creationDate.After(*i.completionDate) {
			i.creationDate = i.completionDate
		}
	})
}

func (i *Item) MarkUndone() error {
	return i.modify(func() {
		i.done = false
		i.completionDate = nil
	})
}

func (i *Item) PrioritizeAs(prio Priority) error {
	return i.modify(func() {
		i.done = false
		i.prio = prio
	})
}

func (i *Item) EditDescription(desc string) error {
	return i.modify(func() {
		i.description = desc
	})
}

func (i *Item) String() string {
	if out, err := DefaultEncoder.encodeItem(i); err == nil {
		return out
	}
	return fmt.Sprintf("%#v", i)
}

func (i *Item) modify(modification func()) error {
	previous := *i
	modification()
	err := i.emit(ModEvent{
		Previous: &previous,
		Current:  i,
	})
	if err != nil {
		*i = previous
	}
	return err
}

func (i *Item) emit(event Event) error {
	if i.emitFunc != nil {
		return i.emitFunc(event)
	}
	return nil
}

func (i *Item) validate() error {
	if err := i.emit(ValidationEvent{i}); err != nil {
		return err
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
