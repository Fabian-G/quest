package todotxt

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const dateLength = 10

var doneRegex = regexp.MustCompile("^x([[:space:]]|$)")
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

type stateFunc func(lexer *parser) (stateFunc, error)

type parser struct {
	remainingLine string
	builderParts  []BuildFunc
}

func (p *parser) Advance(bytes int) {
	p.remainingLine = p.remainingLine[bytes:]
}

func (p *parser) SkipWhiteSpace() {
	p.remainingLine = strings.TrimSpace(p.remainingLine)
}

func ParseItem(todoItem string) (*Item, error) {
	parser := parser{
		remainingLine: todoItem,
		builderParts:  make([]BuildFunc, 0),
	}

	state := doneMarker
	var err error
	for state != nil {
		state, err = state(&parser)
		if err != nil {
			return nil, err
		}
	}

	item, err := BuildItem(parser.builderParts...)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	return item, nil
}

func doneMarker(p *parser) (stateFunc, error) {
	if doneRegex.MatchString(p.remainingLine) {
		p.builderParts = append(p.builderParts, WithDone(true))
		p.Advance(1)
		p.SkipWhiteSpace()
	} else if startsWithSpace(p.remainingLine) {
		p.SkipWhiteSpace()
		return description, nil
	}
	return prio, nil
}

func prio(p *parser) (stateFunc, error) {
	if prioRegex.MatchString(p.remainingLine) {
		prio, err := PriorityFromString(p.remainingLine[:3])
		if err != nil {
			return nil, ParseError{p.remainingLine[:3], err}
		}
		p.builderParts = append(p.builderParts, WithPriority(prio))
		p.Advance(3)
		p.SkipWhiteSpace()
	}
	return dates, nil
}

func dates(p *parser) (stateFunc, error) {
	if twoDatesRegex.MatchString(p.remainingLine) {
		return completionDate, nil
	} else if dateRegex.MatchString(p.remainingLine) {
		return creationDate, nil
	}
	return description, nil
}

func completionDate(p *parser) (stateFunc, error) {
	dateString := p.remainingLine[:dateLength]
	completionDate, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		return nil, ParseError{dateString, err}
	}
	p.builderParts = append(p.builderParts, WithCompletionDate(&completionDate))
	p.Advance(len(dateString))
	p.SkipWhiteSpace()
	return creationDate, nil
}

func creationDate(p *parser) (stateFunc, error) {
	dateString := p.remainingLine[:dateLength]
	creationDate, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		return nil, ParseError{dateString, err}
	}
	p.builderParts = append(p.builderParts, WithCreationDate(&creationDate))
	p.Advance(len(dateString))
	p.SkipWhiteSpace()
	return description, nil
}

func description(p *parser) (stateFunc, error) {
	desc := p.remainingLine
	// Interpret new lines
	desc = strings.ReplaceAll(desc, "\\n", "\n")
	desc = strings.ReplaceAll(desc, "\\r", "\r")
	p.builderParts = append(p.builderParts, WithDescription(desc))
	return nil, nil
}

func startsWithSpace(s string) bool {
	for _, c := range s {
		return unicode.IsSpace(c)
	}
	return false
}
