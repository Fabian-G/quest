package todotxt

import "errors"

var ErrAbort = errors.New("operation aborted by hook")

// ModEvent is an event that is issued whenever a modification to the list happens.
// In case a task was added previous will be nil and current non-nil
// In case an existing task was modified both will be non-nil
// In case a task was removed previous will be non-nil and current will be nil
// It is legal to modify Current from within a hook. e.g. to undo a change do `*Current = *Previous`
type ModEvent struct {
	Previous *Item
	Current  *Item
}

type Hook interface {
	Handle(event ModEvent) error
}

type HookFunc func(event ModEvent) error

func (h HookFunc) Handle(event ModEvent) error {
	return h(event)
}
