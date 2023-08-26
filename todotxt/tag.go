package todotxt

import "regexp"

var tagRegex = regexp.MustCompile("(^| )[^[:space:]:@+]*:[^[:space:]]+( |$)")

type Tags map[string][]string
