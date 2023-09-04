package todotxt

import (
	"slices"
)

type List struct {
	tasks         []*Item
	hooksDisabled bool
	hooks         []Hook
}

func ListOf(items ...*Item) *List {
	newList := &List{}
	for _, item := range items {
		newList.Add(item)
	}
	return newList
}

func (l *List) Tasks() []*Item {
	return l.tasks
}

func (l *List) Add(item *Item) error {
	item.emitFunc = l.emit
	l.tasks = append(l.tasks, item)
	err := l.emit(ModEvent{Previous: nil, Current: item})
	if err != nil {
		l.tasks = l.tasks[:len(l.tasks)-1]
	}
	return err
}
func (l *List) AddHook(hook Hook) {
	l.hooks = append(l.hooks, hook)
}

func (l *List) Remove(idx int) error {
	itemToRemove := l.Get(idx)
	itemToRemove.emitFunc = nil
	l.tasks = slices.Delete(l.tasks, idx, idx+1)
	err := l.emit(ModEvent{Previous: itemToRemove, Current: nil})
	if err != nil {
		l.tasks = slices.Insert(l.tasks, idx, itemToRemove)
	}
	return err
}

func (l *List) Get(idx int) *Item {
	return l.tasks[idx]
}

func (l *List) Len() int {
	return len(l.tasks)
}

func (l *List) emit(me ModEvent) error {
	if l.hooksDisabled {
		return nil
	}

	l.hooksDisabled = true
	defer func() {
		l.hooksDisabled = false
	}()
	for _, h := range l.hooks {
		if err := h.Handle(me); err != nil {
			return err
		}
	}
	return nil
}
