package todotxt

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Tests whether or not a description needs a leading space when formatted to avoid being ambiguous.
// For example when serializing the description "x test" it would be deserialized to {done: true, desc: test} unless
// we add an additional space when serializing: " x test"
var leadingSpaceNeeded = regexp.MustCompile(`^x |^[0-9]{4}-[0-9]{2}-[0-9]{2} |^\([A-Z]\) `)

var DefaultEncoder = Encoder{}

type Encoder struct {
}

func (f Encoder) Encode(w io.Writer, tasks []*Item) error {
	out := bufio.NewWriter(w)
	for _, item := range tasks {
		formattedItem, err := f.encodeItem(item)
		if err != nil {
			return fmt.Errorf("could not marshal item: %w", err)
		}
		if _, err := out.WriteString(formattedItem + "\n"); err != nil {
			return fmt.Errorf("could not write item %v: %w", item, err)
		}
	}
	return out.Flush()
}

func (e Encoder) encodeItem(i *Item) (string, error) {
	if err := i.validate(); err != nil {
		return "", fmt.Errorf("can not format an invalid item: %w", err)
	}
	builder := strings.Builder{}
	if i.Done() {
		builder.WriteString("x")
		builder.WriteString(" ")
	}
	if i.Priority() != PrioNone {
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
	desc := i.description
	desc = strings.TrimSpace(desc)
	desc = strings.ReplaceAll(desc, "\n", "\\n")
	desc = strings.ReplaceAll(desc, "\r", "\\r")
	builder.WriteString(desc)
	return strings.TrimRightFunc(builder.String(), unicode.IsSpace), nil
}
