package todotxt

import (
	"fmt"
	"regexp"
)

type Project string

// Matcher returns a regular expression matching this specific project.
// The Match will contain a leading a trailing whitespace, the project description without
// whitespace will be matched by submatch 1
func (p Project) Matcher() *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf("(?:^| )+(%s)(?: |$)", regexp.QuoteMeta(string(p))))
}

func (p Project) String() string {
	return fmt.Sprintf("+%s", string(p))
}
