package tfmt

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/Fabian-G/todotxt/todotxt"
)

// Tests whether or not a description needs a leading space when formatted to avoid being ambiguous.
// For example when serializing the description "x test" it would be deserialized to {done: true, desc: test} unless
// we add an additional space when serializing: " x test"
var leadingSpaceNeeded = regexp.MustCompile("^x |^[0-9]{4}-[0-9]{2}-[0-9]{2} |^\\([A-Z]\\) ")

var DefaultFormatter = Formatter{}

type Formatter struct {
}

// Format formats an item according to the todotxt spec
func (f *Formatter) Format(i *todotxt.Item) string {
	if i.CompletionDate() != nil && i.CreationDate() == nil {
		// In fact this can not really happen
		panic(errors.New("trying to serialize invalid task. CompletionDate set, but CreationDate is not"))
	}
	if i.Description() == "" {
		return ""
	}
	builder := strings.Builder{}
	if i.Done() {
		builder.WriteString("x")
		builder.WriteString(" ")
	}
	if i.Priority() != todotxt.PrioNone {
		builder.WriteString(i.Priority().String())
		builder.WriteString(" ")
	}
	if i.CompletionDate() != nil {
		builder.WriteString(i.CompletionDate().Format(time.DateOnly))
		builder.WriteString(" ")
	}
	if i.CreationDate() != nil {
		builder.WriteString(i.CreationDate().Format(time.DateOnly))
		builder.WriteString(" ")
	}
	if builder.Len() == 0 && leadingSpaceNeeded.MatchString(i.Description()) {
		builder.WriteString(" ")
	}
	builder.WriteString(i.Description())
	return builder.String()
}
