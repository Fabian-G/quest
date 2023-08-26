package todotxt

import "regexp"

var projectRegex = regexp.MustCompile("(^| )\\+[^[:space:]]+( |$)")

type Project string
