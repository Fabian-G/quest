package todotxt

import (
	"errors"
	"fmt"
	"slices"
)

type ValidationError struct {
	Base          error
	OffendingItem *Item
	ItemIndex     int
}

func (v ValidationError) Unwrap() error {
	return v.Base
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("the item with index %d is invalid: %v\n\tthe offending item was %v", v.ItemIndex, v.Base, v.OffendingItem)
}

type List struct {
	diskOrder     []*Item // The order in which the items are stored on disk
	idxOrder      []*Item // The order according to the default order
	IdxOrderFunc  func(*Item, *Item) int
	hooksDisabled bool
	hooks         []Hook
}

func ListOf(items ...*Item) *List {
	newList := &List{}
	_ = newList.Add(items...) // error can be ignored here, because list does not have any hooks yet
	return newList
}

// Tasks returns the list as a slice of Items ordered like they are ordered on disk.
func (l *List) Tasks() []*Item {
	tasks := make([]*Item, 0, len(l.diskOrder))
	for _, i := range l.diskOrder {
		if i != nil {
			tasks = append(tasks, i)
		}
	}
	return tasks
}

func (l *List) Add(items ...*Item) (err error) {
	originalLength := len(l.diskOrder)
	defer func() {
		if err != nil {
			l.diskOrder = l.diskOrder[:originalLength]
			l.idxOrder = l.idxOrder[:originalLength]
		}
	}()
	for _, i := range items {
		i.emitFunc = l.emit
		l.diskOrder = append(l.diskOrder, i)
		l.idxOrder = append(l.idxOrder, i)
		err = i.validate()
		if err != nil {
			return
		}
		err = l.emit(ModEvent{Previous: nil, Current: i})
		if err != nil {
			return
		}
	}
	return
}

func (l *List) Reindex() {
	if l.IdxOrderFunc != nil {
		slices.SortStableFunc(l.idxOrder, l.IdxOrderFunc)
	}
}

func (l *List) AddHook(hook Hook) {
	l.hooks = append(l.hooks, hook)
}

func (l *List) Remove(idx int) error {
	itemToRemove := l.Get(idx)
	if itemToRemove == nil {
		return fmt.Errorf("item with idx %d does not exists", idx)
	}
	itemToRemove.emitFunc = nil

	diskIdx := slices.Index(l.diskOrder, l.idxOrder[idx])
	l.diskOrder[diskIdx] = nil
	l.idxOrder[idx] = nil
	err := l.emit(ModEvent{Previous: itemToRemove, Current: nil})
	if err != nil {
		l.diskOrder[diskIdx] = itemToRemove
		l.idxOrder[idx] = itemToRemove
	}
	return err
}

func (l *List) Get(idx int) *Item {
	return l.idxOrder[idx]
}

func (l *List) IndexOf(i *Item) int {
	return slices.Index(l.idxOrder, i)
}

func (l *List) Len() int {
	return len(l.Tasks())
}

func (l *List) AllProjects() []Project {
	allProjectsM := make(map[Project]struct{})
	for _, i := range l.idxOrder {
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
	for _, i := range l.idxOrder {
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
	for _, i := range l.idxOrder {
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

func (l *List) validate() error {
	errs := make([]error, 0)
	for key, value := range l.idxOrder {
		baseErr := value.validate()
		if baseErr != nil {
			errs = append(errs, ValidationError{
				Base:          baseErr,
				OffendingItem: value,
				ItemIndex:     key,
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
