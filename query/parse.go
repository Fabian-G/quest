package query

import (
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
		return root.eval(universe, alpha)
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
		return nil, fmt.Errorf("expected identifier, got: %d", id.typ)
	}
	c := p.next()
	if c.typ != itemColon {
		return nil, fmt.Errorf("expected colon, got: %d", id.typ)
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
	return &impl{
		leftChild:  &boolConst{"true"},
		rightChild: &boolConst{"true"},
	}, nil
}
