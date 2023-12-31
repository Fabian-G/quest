package qselect

import (
	"errors"
	"fmt"
	"slices"
)

func parseQQLTree(query string, expectedFreeVars idSet, expectedResultType DType) (node, error) {
	parser := parser{
		lex: lex(query),
	}
	parser.next()
	if parser.lookAhead().typ == eof {
		return &boolConst{val: "true"}, nil
	}
	root, err := parser.parseExp()
	if err != nil {
		return nil, err
	}
	if parser.lookAhead().typ != eof {
		return nil, fmt.Errorf("garbage at the end of expression: %s", parser.lookAhead().val)
	}
	t, err := root.validate(expectedFreeVars)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	if t != expectedResultType {
		return nil, fmt.Errorf("query result must be %s, got: %s", expectedResultType, t)
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
	if in := p.next(); in.typ != itemIn {
		return nil, fmt.Errorf("expected \"in\", got: \"%s\" at position %d", in.val, id.pos)
	}
	collection, err := p.parseExp()
	if err != nil {
		return nil, err
	}
	if c := p.next(); c.typ != itemColon {
		return nil, fmt.Errorf("expected colon, got: \"%s\" at position %d", c.val, c.pos)
	}
	child, err := p.parseExp()
	if err != nil {
		return nil, err
	}

	switch next.typ {
	case itemAllQuant:
		return &allQuant{boundId: id.val, collection: collection, child: child}, nil
	case itemExistQuant:
		return &existQuant{boundId: id.val, collection: collection, child: child}, nil

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
		rightChild, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		next = p.lookAhead()
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
		rightChild, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		next = p.lookAhead()
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
		rightChild, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		next = p.lookAhead()
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
		return p.parseComparison()
	}

	p.next()
	child, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	return &not{
		child: child,
	}, nil
}

func (p *parser) parseComparison() (node, error) {
	comparisons := []itemType{itemEq, itemLeq, itemLt, itemGeq, itemGt}
	child, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}
	for next := p.lookAhead(); slices.Contains(comparisons, next.typ); {
		p.next()
		rightChild, err := p.parseAddSub()
		if err != nil {
			return nil, err
		}
		child = &comparison{
			comparator: next.typ,
			leftChild:  child,
			rightChild: rightChild,
		}
		next = p.lookAhead()
	}
	return child, nil
}

func (p *parser) parseAddSub() (node, error) {
	child, err := p.parseSign()
	if err != nil {
		return nil, err
	}

	for next := p.lookAhead(); next.typ == itemPlus || next.typ == itemMinus; {
		p.next()
		rightChild, err := p.parseSign()
		if err != nil {
			return nil, err
		}
		switch next.typ {
		case itemPlus:
			child = &plus{
				leftChild:  child,
				rightChild: rightChild,
			}
		case itemMinus:
			child = &minus{
				leftChild:  child,
				rightChild: rightChild,
			}
		}
		next = p.lookAhead()
	}
	return child, nil
}

func (p *parser) parseSign() (node, error) {
	next := p.lookAhead()
	if next.typ != itemPlus && next.typ != itemMinus {
		return p.parsePrimary()
	}

	p.next()
	child, err := p.parseSign()
	if err != nil {
		return nil, err
	}

	if next.typ == itemPlus {
		return child, nil
	}
	return &negativeSign{
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
	case itemDuration:
		p.next()
		return &durationConst{
			val: next.val,
		}, nil
	case itemBool:
		p.next()
		return &boolConst{
			val: next.val,
		}, nil
	case itemProjMatch:
		p.next()
		return p.buildProjMatcher(next.val)
	case itemCtxMatch:
		p.next()
		return p.buildCtxMatcher(next.val)
	case itemLeftParen:
		p.next()
		child, err := p.parseExp()
		if err != nil {
			return nil, err
		}
		next := p.next()
		if next.typ != itemRightParen {
			return nil, errors.New("missing closing parenthesis")
		}
		return child, err
	case itemAllQuant:
		return p.parseExp()
	case itemExistQuant:
		return p.parseExp()
	case itemIdent:
		return p.parseIdentifierOrCall()
	default:
		return nil, fmt.Errorf("unexpected token: \"%s\" at position %d", next.val, next.pos)
	}
}

func (p *parser) parseIdentifierOrCall() (node, error) {
	next := p.next()
	fCallTest := p.lookAhead()
	if fCallTest.typ == itemLeftParen {
		args, err := p.parseArgs()
		if err != nil {
			return nil, err
		}
		return &call{
			name: next.val,
			args: args,
			fn:   functions[next.val],
		}, nil
	} else if fn, ok := functions[next.val]; ok { // This is really a function call without args
		return &call{
			name: next.val,
			fn:   fn,
			args: &args{},
			ifBound: &identifier{ // Except if this id is bound by a quantifier. In that cas treat it as an identifier
				name: next.val,
			},
		}, nil
	}
	return &identifier{
		name: next.val,
	}, nil
}

func (p *parser) parseArgs() (*args, error) {
	var arguments []node
	lParen := p.next()
	if lParen.typ != itemLeftParen {
		return nil, fmt.Errorf("expected opening parenthesis got: \"%s\" at position %d", lParen.val, lParen.pos)
	}
	if p.lookAhead().typ == itemRightParen {
		p.next()
		return &args{}, nil // args list is empty
	}
	for {
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

func (p *parser) buildProjMatcher(proj string) (node, error) {
	query := fmt.Sprintf("(exists p in projects(it): dotPrefix(p, \"%s\"))", proj)
	parser := *p
	parser.lex = lex(query)
	parser.next()
	tree, err := parser.parseExp()
	return tree, err
}

func (p *parser) buildCtxMatcher(ctx string) (node, error) {
	query := fmt.Sprintf("(exists p in contexts(it): dotPrefix(p, \"%s\"))", ctx)
	parser := *p
	parser.lex = lex(query)
	parser.next()
	tree, err := parser.parseExp()
	return tree, err
}
