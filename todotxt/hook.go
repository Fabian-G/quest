package todotxt

import "errors"

var ErrAbort = errors.New("operation aborted by hook")

type Event interface {
	Dispatch(*List, Hook) error
}

// ModEvent is an event that is issued whenever a modification to the list happens.
// In case a task was added previous will be nil and current non-nil
// In case an existing task was modified both will be non-nil
// In case a task was removed previous will be non-nil and current will be nil
// It is legal to modify Current from within a hook. e.g. to undo a change do `*Current = *Previous`
type ModEvent struct {
	Previous *Item
	Current  *Item
}

func (m ModEvent) IsCompleteEvent() bool {
	return m.Previous != nil && m.Current != nil && !m.Previous.Done() && m.Current.Done()
}

func (m ModEvent) Dispatch(list *List, h Hook) error {
	return h.OnMod(list, m)
}

type ValidationEvent struct {
	Item *Item
}

func (v ValidationEvent) Dispatch(list *List, h Hook) error {
	return h.OnValidate(list, v)
}

type Hook interface {
	OnMod(list *List, event ModEvent) error
	OnValidate(list *List, event ValidationEvent) error
}

type HookFunc func(*List, ModEvent) error

func (h HookFunc) OnMod(list *List, event ModEvent) error {
	return h(list, event)
}

func (h HookFunc) OnValidate(*List, ValidationEvent) error {
	return nil
}
