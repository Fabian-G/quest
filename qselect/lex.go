// Highly inspired by text/template
package qselect

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type item struct {
	typ itemType
	pos int
	val string
}

type itemType int

const (
	itemErr itemType = iota
	itemIdent
	itemAnd
	itemEq
	itemOr
	itemImpl
	itemLt
	itemGt
	itemLeq
	itemGeq
	itemNot
	itemString
	itemInt
	itemDuration
	itemBool
	itemAllQuant
	itemIn
	itemExistQuant
	itemLeftParen
	itemRightParen
	itemComma
	itemColon
	itemProjMatch
	itemCtxMatch
	itemPlus
	itemMinus
)

const eof = -1

type stateFunc func(*lexer) stateFunc

type lexer struct {
	input      string
	pos        int
	start      int
	atEOF      bool
	item       item
	parenDepth int
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.atEOF = true
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += w
	return r
}

func (l *lexer) peek() rune {
	next := l.next()
	l.backup()
	return next
}

func (l *lexer) backup() {
	if !l.atEOF && l.pos > 0 {
		_, w := utf8.DecodeLastRuneInString(l.input[:l.pos])
		l.pos -= w
	}
}

func (l *lexer) thisItem(typ itemType) item {
	item := item{
		typ: typ,
		pos: l.start,
		val: l.input[l.start:l.pos],
	}
	l.start = l.pos
	return item
}

func (l *lexer) emitItem(item item) stateFunc {
	l.item = item
	return nil
}

func (l *lexer) emit(t itemType) stateFunc {
	return l.emitItem(l.thisItem(t))
}

func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for l.accept(valid) {
	}
}

func (l *lexer) discard() {
	l.start = l.pos
}

func (l *lexer) reset() {
	l.pos = l.start
}

func (l *lexer) errorf(format string, args ...any) stateFunc {
	l.item = item{
		typ: itemErr,
		pos: l.start,
		val: fmt.Sprintf(format, args...),
	}
	l.start = 0
	l.pos = 0
	l.input = l.input[:0]
	return nil
}

func (l *lexer) nextItem() item {
	l.item = item{
		typ: eof,
		pos: l.pos,
		val: "EOF",
	}
	state := lexQuery
	for {
		state = state(l)
		if state == nil {
			return l.item
		}
	}
}

func lex(query string) *lexer {
	lex := &lexer{
		input: query,
	}
	return lex
}

func lexQuery(l *lexer) stateFunc {
	switch r := l.next(); {
	case r == eof:
		if l.parenDepth > 0 {
			return l.errorf("missing closing paren")
		}
		return nil
	case unicode.IsSpace(r):
		l.discard()
		return lexQuery
	case r == '+':
		if unicode.IsLetter(l.peek()) {
			return lexProjectMatcherOrPlus
		}
		return l.emit(itemPlus)
	case r == '@':
		return lexContextMatcher
	case r == '!':
		return l.emit(itemNot)
	case r == '"':
		return lexStringLiteral
	case r == '&':
		if l.next() != '&' {
			return l.errorf("expected &&")
		}
		return l.emit(itemAnd)
	case r == '<':
		if l.next() == '=' {
			return l.emit(itemLeq)
		}
		l.backup()
		return l.emit(itemLt)
	case r == '>':
		if l.next() == '=' {
			return l.emit(itemGeq)
		}
		l.backup()
		return l.emit(itemGt)
	case r == '=':
		if l.next() != '=' {
			return l.errorf("expected ==")
		}
		return l.emit(itemEq)
	case r == '|':
		if l.next() != '|' {
			return l.errorf("expected ||")
		}
		return l.emit(itemOr)
	case r == '-':
		if l.peek() != '>' {
			return l.emit(itemMinus)
		}
		l.accept(">")
		return l.emit(itemImpl)
	case r == ',':
		return l.emit(itemComma)
	case r == '(' || r == '{':
		l.parenDepth++
		return l.emit(itemLeftParen)
	case r == ')' || r == '}':
		l.parenDepth--
		if l.parenDepth < 0 {
			return l.errorf("unexpected right paren")
		}
		return l.emit(itemRightParen)
	case r == ':':
		return l.emit(itemColon)
	case r == '-' || ('0' <= r && r <= '9'):
		l.backup()
		return lexIntOrDuration
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier
	default:
		return l.errorf("unexpected character %#U", r)
	}
}

func lexIdentifier(l *lexer) stateFunc {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):

		default:
			l.backup()
			if !l.isAtTerminator() {
				return l.errorf("unexpected token in identifier %#U", r)
			}

			word := l.input[l.start:l.pos]
			if word == "exists" && l.accept(" ") {
				return l.emit(itemExistQuant)
			}
			if word == "forall" && l.accept(" ") {
				return l.emit(itemAllQuant)
			}
			if word == "true" || word == "false" {
				return l.emit(itemBool)
			}
			if word == "and" {
				return l.emit(itemAnd)
			}
			if word == "or" {
				return l.emit(itemOr)
			}
			if word == "impl" {
				return l.emit(itemImpl)
			}
			if word == "in" {
				return l.emit(itemIn)
			}
			return l.emit(itemIdent)
		}
	}
}

func lexProjectMatcherOrPlus(l *lexer) stateFunc {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):

		default:
			l.backup()
			if !l.isAtTerminator() || (l.pos-l.start) == 1 {
				l.reset()
				l.accept("+")
				return l.emit(itemPlus)
			}

			return l.emit(itemProjMatch)
		}
	}
}

func lexContextMatcher(l *lexer) stateFunc {
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):

		default:
			l.backup()
			if !l.isAtTerminator() {
				return l.errorf("unexpected token in context matcher %#U", r)
			}

			return l.emit(itemCtxMatch)
		}
	}
}

func lexIntOrDuration(l *lexer) stateFunc {
	digits := "0123456789"
	l.acceptRun(digits)
	duration := l.accept("dwmy")
	if isAlphaNumeric(l.peek()) {
		return l.errorf("Unexpected character at end of integer")
	}
	if duration {
		return l.emit(itemDuration)
	}
	return l.emit(itemInt)
}

func lexStringLiteral(l *lexer) stateFunc {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof {
				break
			}
			fallthrough
		case eof:
			return l.errorf("unterminated quoted string")
		case '"':
			break Loop
		}
	}
	return l.emit(itemString)
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func (l *lexer) isAtTerminator() bool {
	r := l.peek()
	if unicode.IsSpace(r) {
		return true
	}
	switch r {
	case eof, ',', ')', '(', ':', '{', '}', '|', '&', '-', '+', '=', '>', '<':
		return true
	}
	return false
}
