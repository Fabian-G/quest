package todotxt

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// Tests whether or not a description needs a leading space when formatted to avoid being ambiguous.
// For example when serializing the description "x test" it would be deserialized to {done: true, desc: test} unless
// we add an additional space when serializing: " x test"
var leadingSpaceNeeded = regexp.MustCompile("^x |^[0-9]{4}-[0-9]{2}-[0-9]{2} |^\\([A-Z]\\) ")

var DefaultFormatter = Formatter{}

type Formatter struct {
}

// Format formats an item according to the todotxt spec
func (f *Formatter) Format(i *Item) string {
	if i.completionDate != zeroTime && i.creationDate == zeroTime {
		// In fact this can not really happen
		panic(errors.New("trying to serialize invalid task. CompletionDate set, but CreationDate is not"))
	}
	if i.description == "" {
		return ""
	}
	builder := strings.Builder{}
	if i.Done() {
		builder.WriteString("x")
		builder.WriteString(" ")
	}
	if i.Priority() != PrioNone {
		builder.WriteString(i.prio.String())
		builder.WriteString(" ")
	}
	if i.completionDate != zeroTime {
		builder.WriteString(i.completionDate.Format(time.DateOnly))
		builder.WriteString(" ")
	}
	if i.creationDate != zeroTime {
		builder.WriteString(i.creationDate.Format(time.DateOnly))
		builder.WriteString(" ")
	}
	if builder.Len() == 0 && leadingSpaceNeeded.MatchString(i.description) {
		builder.WriteString(" ")
	}
	builder.WriteString(i.description)
	return builder.String()
}
