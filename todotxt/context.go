package todotxt

import "regexp"

var contextRegex = regexp.MustCompile("(^| )@[^[:space:]]+( |$)")

type Context string
