package parse

import (
	"errors"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/Fabian-G/todotxt/todotxt"
)

const dateLength = 10

var doneRegex = regexp.MustCompile("^x( |$)")
var prioRegex = regexp.MustCompile("^\\([A-Z]\\)([[:space:]]|$)")
var dateRegex = regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}([[:space:]]|$)")
var twoDatesRegex = regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}[[:space:]]+[0-9]{4}-[0-9]{2}-[0-9]{2}([[:space:]]|$)")

type stateFunc func(lexer *lexer) stateFunc

var ErrEmpty error = errors.New("the todo item is empty")

type lexer struct {
	remainingLine string
	builderParts  []todotxt.BuildFunc
}

func (lex *lexer) Advance(bytes int) {
	lex.remainingLine = lex.remainingLine[bytes:]
}

func (lex *lexer) SkipWhiteSpace() {
	lex.remainingLine = strings.TrimSpace(lex.remainingLine)
}

func Item(todoItem string) (*todotxt.Item, error) {
	if len(strings.TrimSpace(todoItem)) == 0 {
		return nil, ErrEmpty
	}

	lexer := lexer{
		remainingLine: todoItem,
		builderParts:  make([]todotxt.BuildFunc, 0),
	}

	state := doneMarker
	for state != nil {
		state = state(&lexer)
	}

	return todotxt.Build(lexer.builderParts...)
}

func doneMarker(lex *lexer) stateFunc {
	if doneRegex.MatchString(lex.remainingLine) {
		lex.builderParts = append(lex.builderParts, todotxt.WithDone(true))
		lex.Advance(1)
		lex.SkipWhiteSpace()
	} else if startsWithSpace(lex.remainingLine) {
		lex.SkipWhiteSpace()
		return description
	}
	return prio
}

func prio(lex *lexer) stateFunc {
	if prioRegex.MatchString(lex.remainingLine) {
		prio, err := todotxt.PriorityFromString(lex.remainingLine[:3])
		if err != nil {
			log.Printf("Warning: encountered something that looks like a priority, but is not (%s) treating as description", prio)
			return description
		}
		lex.builderParts = append(lex.builderParts, todotxt.WithPriority(prio))
		lex.Advance(3)
		lex.SkipWhiteSpace()
	}
	return dates
}

func dates(lex *lexer) stateFunc {
	if twoDatesRegex.MatchString(lex.remainingLine) {
		return completionDate
	} else if dateRegex.MatchString(lex.remainingLine) {
		return creationDate
	}
	return description
}

func completionDate(lex *lexer) stateFunc {
	dateString := lex.remainingLine[:dateLength]
	completionDate, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		log.Printf("Warning: Encountered a string that looks like a date, but is not (%s) treating as description", dateString)
		return description
	}
	lex.builderParts = append(lex.builderParts, todotxt.WithCompletionDate(&completionDate))
	lex.Advance(len(dateString))
	lex.SkipWhiteSpace()
	return creationDate
}

func creationDate(lex *lexer) stateFunc {
	dateString := lex.remainingLine[:dateLength]
	creationDate, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		log.Printf("Warning: Encountered a string that looks like a date, but is not (%s) treating as description", dateString)
		return description
	}
	lex.builderParts = append(lex.builderParts, todotxt.WithCreationDate(&creationDate))
	lex.Advance(len(dateString))
	lex.SkipWhiteSpace()
	return description
}

func description(lex *lexer) stateFunc {
	lex.builderParts = append(lex.builderParts, todotxt.WithDescription(lex.remainingLine))
	return nil
}

func startsWithSpace(s string) bool {
	for _, c := range s {
		return unicode.IsSpace(c)
	}
	return false
}
