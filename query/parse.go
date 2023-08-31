package query

import (
	"errors"
	"fmt"

	"github.com/Fabian-G/todotxt/todotxt"
)

type Query func(todotxt.List, *todotxt.Item) bool

func Compile(query string) (Query, error) {
	root, err := parseTree(query)
	if err != nil {
		return nil, err
	}
	evalFunc := func(universe todotxt.List, it *todotxt.Item) bool {
		alpha := make(map[string]*todotxt.Item)
		alpha["it"] = it
		return root.eval(universe, alpha).(bool)
	}
	return evalFunc, nil
}

func parseTree(query string) (node, error) {
	parser := parser{
		lex: lex(query),
	}
	parser.next()
	root, err := parser.parseExp()
	if err != nil {
		return nil, err
	}
	t, err := root.validate()
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if t != qBool {
		return nil, fmt.Errorf("query result must be bool, got: %s", t)
	}
	return root, nil
}

type parser struct {
	lex          *lexer
	currentToken item
}

func (p *parser) lookAhead() item {
	return p.currentToken
}

func (p *parser) next() item {
	tmp := p.currentToken
	p.currentToken = p.lex.nextItem()
	return tmp
}

func (p *parser) parseExp() (node, error) {
	return p.parseQuant()
}

func (p *parser) parseQuant() (node, error) {
	next := p.lookAhead()
	if next.typ != itemAllQuant && next.typ != itemExistQuant {
		return p.parseImpl()
	}

	p.next()
	id := p.next()
	if id.typ != itemIdent {
		return nil, fmt.Errorf("expected identifier, got: \"%s\" at position %d", id.val, id.pos)
	}
	c := p.next()
	if c.typ != itemColon {
		return nil, fmt.Errorf("expected colon, got: \"%s\" at position %d", id.val, id.pos)
	}
	child, err := p.parseExp()
	if err != nil {
		return nil, err
	}

	switch next.typ {
	case itemAllQuant:
		return &allQuant{boundId: id.val, child: child}, nil
	case itemExistQuant:
		return &existQuant{boundId: id.val, child: child}, nil

	}
	panic("This statement can not be reached")
}

func (p *parser) parseImpl() (node, error) {
	child, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	for next := p.lookAhead(); next.typ == itemImpl; {
		p.next()
		next = p.lookAhead()
		rightChild, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		child = &impl{
			leftChild:  child,
			rightChild: rightChild,
		}
	}
	return child, nil
}

func (p *parser) parseOr() (node, error) {
	child, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for next := p.lookAhead(); next.typ == itemOr; {
		p.next()
		next = p.lookAhead()
		rightChild, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		child = &or{
			leftChild:  child,
			rightChild: rightChild,
		}
	}
	return child, nil
}

func (p *parser) parseAnd() (node, error) {
	child, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for next := p.lookAhead(); next.typ == itemAnd; {
		p.next()
		next = p.lookAhead()
		rightChild, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		child = &and{
			leftChild:  child,
			rightChild: rightChild,
		}
	}
	return child, nil

}

func (p *parser) parseNot() (node, error) {
	next := p.lookAhead()
	if next.typ != itemNot {
		return p.parsePrimary()
	}

	p.next()
	child, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	return &not{
		child: child,
	}, nil
}

func (p *parser) parsePrimary() (node, error) {
	next := p.lookAhead()
	switch next.typ {
	case itemString:
		p.next()
		return &stringConst{
			val: next.val,
		}, nil
	case itemInt:
		p.next()
		return &intConst{
			val: next.val,
		}, nil
	case itemBool:
		p.next()
		return &boolConst{
			val: next.val,
		}, nil
	case itemLeftParen:
		p.next()
		child, err := p.parseExp()
		next := p.next()
		if next.typ != itemRightParen {
			return nil, errors.New("missing closing parenthesis")
		}
		return child, err
	case itemIdent:
		p.next()
		fCallTest := p.lookAhead()
		if fCallTest.typ == itemLeftParen {
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			return &call{
				name: next.val,
				args: args,
			}, nil
		}
		return &identifier{
			name: next.val,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected token: \"%s\" at position %d", next.val, next.pos)
	}
}

func (p *parser) parseArgs() (*args, error) {
	var arguments []node
	lParen := p.next()
	if lParen.typ != itemLeftParen {
		return nil, fmt.Errorf("expected opening parenthesis got: \"%s\" at position %d", lParen.val, lParen.pos)
	}
	for next := p.lookAhead(); next.typ != itemRightParen; next = p.lookAhead() {
		arg, err := p.parseExp()
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, arg)
		commaOrParen := p.next()
		if commaOrParen.typ == itemComma {
			continue
		}
		if commaOrParen.typ == itemRightParen {
			break
		}
		return nil, fmt.Errorf("expected comma got: \"%s\" at position %d", commaOrParen.val, commaOrParen.pos)
	}
	return &args{
		children: arguments,
	}, nil
}
