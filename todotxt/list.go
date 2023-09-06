package todotxt

import (
	"cmp"
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
	tasks         map[int]*Item // This is a map to make sure indices remain stable between cli invocations
	hooksDisabled bool
	hooks         []Hook
	currentIdx    int
}

func ListOf(items ...*Item) *List {
	newList := &List{}
	for _, item := range items {
		newList.Add(item)
	}
	return newList
}

// Tasks returns the list as a slice of Items ordered by their respective index.
func (l *List) Tasks() []*Item {
	type pair struct {
		key   int
		value *Item
	}
	pairs := make([]pair, 0, len(l.taskMap()))
	for key, value := range l.taskMap() {
		pairs = append(pairs, pair{key, value})
	}
	slices.SortFunc(pairs, func(a, b pair) int { return cmp.Compare(a.key, b.key) })
	items := make([]*Item, 0, len(pairs))
	for _, pair := range pairs {
		items = append(items, pair.value)
	}
	return items
}

func (l *List) Add(item *Item) (err error) {
	item.emitFunc = l.emit
	l.taskMap()[l.currentIdx] = item
	l.currentIdx++
	defer func() {
		if err != nil {
			delete(l.taskMap(), l.currentIdx-1)
			l.currentIdx--
		}
	}()
	err = item.validate()
	if err != nil {
		return
	}
	err = l.emit(ModEvent{Previous: nil, Current: item})
	if err != nil {
		return
	}
	return
}

func (l *List) AddHook(hook Hook) {
	l.hooks = append(l.hooks, hook)
}

func (l *List) Remove(idx int) error {
	itemToRemove := l.Get(idx)
	itemToRemove.emitFunc = nil
	delete(l.taskMap(), idx)
	err := l.emit(ModEvent{Previous: itemToRemove, Current: nil})
	if err != nil {
		l.taskMap()[idx] = itemToRemove
	}
	return err
}

func (l *List) Get(idx int) *Item {
	return l.taskMap()[idx]
}

func (l *List) IndexOf(i *Item) int {
	for key, value := range l.taskMap() {
		if i == value {
			return key
		}
	}
	return -1
}

func (l *List) Len() int {
	return len(l.taskMap())
}

func (l *List) validate() error {
	errs := make([]error, 0)
	for key, value := range l.taskMap() {
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

func (l *List) taskMap() map[int]*Item {
	if l.tasks == nil {
		l.tasks = make(map[int]*Item)
	}
	return l.tasks
}
