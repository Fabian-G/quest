package todotxt

import (
	"fmt"
	"regexp"
)

type Context string

// Matcher returns a regular expression matching this specific context.
// The Match will contain a leading and trailing whitespace, the context description without
// whitespace will be matched by submatch 1
func (c Context) Matcher() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("(?:^| )@(%s)(?: |$)", regexp.QuoteMeta(string(c))))
}

func (c Context) String() string {
	return fmt.Sprintf("@%s", string(c))
}
