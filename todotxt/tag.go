package todotxt

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type Tags map[string][]string

func (t Tags) Keys() []string {
	keys := make([]string, 0, len(t))
	for key := range t {
		keys = append(keys, key)
	}
	return keys
}

// MatcherForTag returns a regular expression that matches a tag with the supplied key.
// To access the value consult submatch 1
//
// Any leading/trailing whitespace in key will be removed before constructing the matcher
// Any remaining whitespace will be replaced with +
func MatcherForTag(key string) *regexp.Regexp {
	key = prepareKey(key)
	return regexp.MustCompile(fmt.Sprintf("(?:^| )%s:([^[:space:]]+)(?: |$)", regexp.QuoteMeta(key)))
}

func prepareKey(key string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '+'
		}
		return r
	}, strings.TrimSpace(key))
}
