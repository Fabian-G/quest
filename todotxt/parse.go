package todotxt

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const dateLength = 10

var doneRegex = regexp.MustCompile("^x( |$)")
var prioRegex = regexp.MustCompile("^\\([A-Z]\\)([[:space:]]|$)")
var dateRegex = regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}([[:space:]]|$)")
var twoDatesRegex = regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}[[:space:]]+[0-9]{4}-[0-9]{2}-[0-9]{2}([[:space:]]|$)")

type ParseError struct {
	subStr  string
	baseErr error
}

func (p ParseError) Error() string {
	return fmt.Sprintf("ParseError at %s: %v", p.subStr, p.baseErr)
}

type stateFunc func(lexer *lexer) (stateFunc, error)

type lexer struct {
	remainingLine string
	builderParts  []BuildFunc
}

func (lex *lexer) Advance(bytes int) {
	lex.remainingLine = lex.remainingLine[bytes:]
}

func (lex *lexer) SkipWhiteSpace() {
	lex.remainingLine = strings.TrimSpace(lex.remainingLine)
}

func ParseItem(todoItem string) (*Item, error) {
	lexer := lexer{
		remainingLine: todoItem,
		builderParts:  make([]BuildFunc, 0),
	}

	state := doneMarker
	var err error
	for state != nil {
		state, err = state(&lexer)
		if err != nil {
			return nil, err
		}
	}

	item, err := Build(lexer.builderParts...)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	return item, nil
}

func doneMarker(lex *lexer) (stateFunc, error) {
	if doneRegex.MatchString(lex.remainingLine) {
		lex.builderParts = append(lex.builderParts, WithDone(true))
		lex.Advance(1)
		lex.SkipWhiteSpace()
	} else if startsWithSpace(lex.remainingLine) {
		lex.SkipWhiteSpace()
		return description, nil
	}
	return prio, nil
}

func prio(lex *lexer) (stateFunc, error) {
	if prioRegex.MatchString(lex.remainingLine) {
		prio, err := PriorityFromString(lex.remainingLine[:3])
		if err != nil {
			return nil, ParseError{lex.remainingLine[:3], err}
		}
		lex.builderParts = append(lex.builderParts, WithPriority(prio))
		lex.Advance(3)
		lex.SkipWhiteSpace()
	}
	return dates, nil
}

func dates(lex *lexer) (stateFunc, error) {
	if twoDatesRegex.MatchString(lex.remainingLine) {
		return completionDate, nil
	} else if dateRegex.MatchString(lex.remainingLine) {
		return creationDate, nil
	}
	return description, nil
}

func completionDate(lex *lexer) (stateFunc, error) {
	dateString := lex.remainingLine[:dateLength]
	completionDate, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		return nil, ParseError{dateString, err}
	}
	lex.builderParts = append(lex.builderParts, WithCompletionDate(&completionDate))
	lex.Advance(len(dateString))
	lex.SkipWhiteSpace()
	return creationDate, nil
}

func creationDate(lex *lexer) (stateFunc, error) {
	dateString := lex.remainingLine[:dateLength]
	creationDate, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		return nil, ParseError{dateString, err}
	}
	lex.builderParts = append(lex.builderParts, WithCreationDate(&creationDate))
	lex.Advance(len(dateString))
	lex.SkipWhiteSpace()
	return description, nil
}

func description(lex *lexer) (stateFunc, error) {
	lex.builderParts = append(lex.builderParts, WithDescription(lex.remainingLine))
	return nil, nil
}

func startsWithSpace(s string) bool {
	for _, c := range s {
		return unicode.IsSpace(c)
	}
	return false
}
