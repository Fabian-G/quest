package hook

import (
	"github.com/Fabian-G/quest/qprojection"
	"github.com/Fabian-G/quest/todotxt"
)

type ClearOnDone struct {
	Clear []string
}

func (c ClearOnDone) OnMod(list *todotxt.List, event todotxt.ModEvent) error {
	if event.IsCompleteEvent() {
		return event.Current.EditDescription(event.Current.CleanDescription(qprojection.ExpandCleanExpression(list, c.Clear)))
	}
	return nil
}

func (c ClearOnDone) OnValidate(list *todotxt.List, event todotxt.ValidationEvent) error {
	return nil
}
