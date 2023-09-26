package todotxt

import (
	"errors"
	"fmt"
	"maps"
	"slices"
)

type ValidationError struct {
	Base          error
	OffendingItem *Item
	Line          int
}

func (v ValidationError) Unwrap() error {
	return v.Base
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("the item on line %d is invalid: %v\n\tthe offending item was \"%s\"", v.Line, v.Base, v.OffendingItem.Description())
}

type listSnapshot struct {
	items     []Item
	deletions map[int]struct{}
}

type List struct {
	snapshot       *listSnapshot
	deletionsStore map[int]struct{}
	items          []*Item // The items (event the deleted items)
	hooksDisabled  bool
	hooks          []Hook
}

func ListOf(items ...*Item) *List {
	newList := &List{}
	_ = newList.Add(items...) // error can be ignored here, because list does not have any hooks yet
	return newList
}

// Tasks returns all non deleted items like they are ordered on disk.
func (l *List) Tasks() []*Item {
	tasks := make([]*Item, 0, len(l.items)-len(l.deletions()))
	for idx, item := range l.items {
		if _, ok := l.deletions()[idx]; !ok {
			tasks = append(tasks, item)
		}
	}
	return tasks
}

func (l *List) Add(items ...*Item) (err error) {
	originalLength := len(l.items)
	defer func() {
		if err != nil {
			l.items = l.items[:originalLength]
		}
	}()
	for _, i := range items {
		i.emitFunc = l.emit
		l.items = append(l.items, i)
		// Order of events is important here, because the ModEvent may transform an invalid task into a valid one
		err = l.emit(ModEvent{Previous: nil, Current: i})
		if err != nil {
			return
		}
		err = i.validate()
		if err != nil {
			return
		}
	}
	return
}

func (l *List) AddHook(hook Hook) {
	l.hooks = append(l.hooks, hook)
}

func (l *List) Remove(idx int) error {
	realIdx := idx - 1
	if realIdx < 0 || realIdx >= len(l.items) {
		return fmt.Errorf("idx %d out of range %d-%d", idx, min(1, len(l.items)), len(l.items))
	}
	l.deletions()[realIdx] = struct{}{}
	err := l.emit(ModEvent{Previous: l.items[realIdx], Current: nil})
	if err != nil {
		delete(l.deletions(), realIdx)
	}
	return err
}

// Get returns the item with the specified idx. The idx is 1-based (see IndexOf)
func (l *List) Get(idx int) *Item {
	return l.items[idx-1]
}

// IndexOf returns the index of the according to the IdxOrderFunc
// The index is 1 bases so that (depending on the IdxOrderFunc) the index corresponds to the line number
func (l *List) IndexOf(i *Item) int {
	return slices.Index(l.items, i) + 1
}

func (l *List) Len() int {
	return len(l.Tasks())
}

func (l *List) AllProjects() []Project {
	allProjectsM := make(map[Project]struct{})
	for _, i := range l.items {
		projects := i.Projects()
		for _, p := range projects {
			allProjectsM[p] = struct{}{}
		}
	}

	allProjects := make([]Project, 0, len(allProjectsM))
	for p := range allProjectsM {
		allProjects = append(allProjects, p)
	}
	return allProjects
}

func (l *List) AllContexts() []Context {
	allContextsM := make(map[Context]struct{})
	for _, i := range l.items {
		contexts := i.Contexts()
		for _, c := range contexts {
			allContextsM[c] = struct{}{}
		}
	}

	allContexts := make([]Context, 0, len(allContextsM))
	for c := range allContextsM {
		allContexts = append(allContexts, c)
	}
	return allContexts
}

func (l *List) AllTags() []string {
	allTagsM := make(map[string]struct{})
	for _, i := range l.items {
		tags := i.Tags()
		for k := range tags {
			allTagsM[k] = struct{}{}
		}
	}

	allTags := make([]string, 0, len(allTagsM))
	for t := range allTagsM {
		allTags = append(allTags, t)
	}
	return allTags
}

func (l *List) Secret(secretChange func() error) error {
	l.hooksDisabled = true
	defer func() {
		l.hooksDisabled = false
	}()
	return secretChange()
}

func (l *List) Snapshot() {
	itemData := make([]Item, 0, len(l.items))
	for _, item := range l.items {
		itemData = append(itemData, *item)
	}
	l.snapshot = &listSnapshot{
		deletions: maps.Clone(l.deletions()),
		items:     itemData,
	}
}

func (l *List) Reset() {
	if l.snapshot == nil {
		return
	}
	l.deletionsStore = maps.Clone(l.snapshot.deletions)
	for i, item := range l.snapshot.items {
		*l.items[i] = item
	}
	l.items = l.items[:len(l.snapshot.items)]
}

func (l *List) deletions() map[int]struct{} {
	if l.deletionsStore == nil {
		l.deletionsStore = make(map[int]struct{})
	}
	return l.deletionsStore
}

func (l *List) validate() error {
	errs := make([]error, 0)
	for _, value := range l.Tasks() {
		baseErr := value.validate()
		if baseErr != nil {
			errs = append(errs, ValidationError{
				Base:          baseErr,
				OffendingItem: value,
				Line:          slices.Index(l.items, value) + 1,
			})
		}
	}
	return errors.Join(errs...)
}

func (l *List) emit(me Event) error {
	if l.hooksDisabled {
		return nil
	}

	l.hooksDisabled = true
	defer func() {
		l.hooksDisabled = false
	}()
	for _, h := range l.hooks {
		if err := me.Dispatch(h); err != nil {
			return err
		}
	}
	return nil
}
