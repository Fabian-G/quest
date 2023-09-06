package todotxt

import "errors"

var ErrAbort = errors.New("operation aborted by hook")

type Event interface {
	Dispatch(Hook) error
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

func (m ModEvent) Dispatch(h Hook) error {
	return h.OnMod(m)
}

type ValidationEvent struct {
	Item *Item
}

func (v ValidationEvent) Dispatch(h Hook) error {
	return h.OnValidate(v)
}

type HookBuilder func(*List) Hook

type Hook interface {
	OnMod(event ModEvent) error
	OnValidate(event ValidationEvent) error
}

type HookFunc func(event ModEvent) error

func (h HookFunc) OnMod(event ModEvent) error {
	return h(event)
}

func (h HookFunc) OnValidate(ValidationEvent) error {
	return nil
}
