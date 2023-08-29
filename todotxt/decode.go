package todotxt

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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

var DefaultDecoder = Decoder{}

type Decoder struct {
}

func (d *Decoder) Decode(r io.Reader) (List, error) {
	list := List(make([]*Item, 0))
	var errs []error
	var lineNumber int
	in := bufio.NewScanner(r)
	for in.Scan() {
		lineNumber++
		text := in.Text()
		item, err := parseItem(text)
		if err != nil {
			errs = append(errs, ReadError{
				BaseError:  err,
				Line:       text,
				LineNumber: lineNumber,
			})
		}
		list = append(list, item)
	}
	if err := in.Err(); err != nil {
		return nil, fmt.Errorf("could not read input: %w", err)
	}
	return list, errors.Join(errs...)
}

type ReadError struct {
	BaseError  error
	Line       string
	LineNumber int
}

func (r ReadError) Error() string {
	return fmt.Sprintf("could not read line number %d: %v\n\tThe problematic line was: %s", r.LineNumber, r.BaseError, r.Line)
}

func (r ReadError) Unwrap() error {
	return r.BaseError
}

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

func (p *parser) advance(bytes int) {
	p.remainingLine = p.remainingLine[bytes:]
}

func (p *parser) skipWhiteSpace() {
	p.remainingLine = strings.TrimSpace(p.remainingLine)
}

func parseItem(todoItem string) (*Item, error) {
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
		p.advance(1)
		p.skipWhiteSpace()
	} else if startsWithSpace(p.remainingLine) {
		p.skipWhiteSpace()
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
		p.advance(3)
		p.skipWhiteSpace()
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
	p.advance(len(dateString))
	p.skipWhiteSpace()
	return creationDate, nil
}

func creationDate(p *parser) (stateFunc, error) {
	dateString := p.remainingLine[:dateLength]
	creationDate, err := time.Parse(time.DateOnly, dateString)
	if err != nil {
		return nil, ParseError{dateString, err}
	}
	p.builderParts = append(p.builderParts, WithCreationDate(&creationDate))
	p.advance(len(dateString))
	p.skipWhiteSpace()
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
