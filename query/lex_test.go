package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_lexOnValidQueries(t *testing.T) {
	testCases := map[string]struct {
		query          string
		expectedTokens []itemType
	}{
		"Single relation symbol": {
			query:          "R(x)",
			expectedTokens: []itemType{itemIdent, itemLeftParen, itemIdent, itemRightParen},
		},
		"Conjunction of two relation symbols": {
			query:          "R(x) && L(y)",
			expectedTokens: []itemType{itemIdent, itemLeftParen, itemIdent, itemRightParen, itemAnd, itemIdent, itemLeftParen, itemIdent, itemRightParen},
		},
		"Disjunction of two relation symbols": {
			query:          "R(x) || L(y)",
			expectedTokens: []itemType{itemIdent, itemLeftParen, itemIdent, itemRightParen, itemOr, itemIdent, itemLeftParen, itemIdent, itemRightParen},
		},
		"Implication between two relation symbols": {
			query:          "R(x) -> L(y)",
			expectedTokens: []itemType{itemIdent, itemLeftParen, itemIdent, itemRightParen, itemImpl, itemIdent, itemLeftParen, itemIdent, itemRightParen},
		},
		"Negation of one relation symbol": {
			query:          "!R(x)",
			expectedTokens: []itemType{itemNot, itemIdent, itemLeftParen, itemIdent, itemRightParen},
		},
		"String literal in relation symbol": {
			query:          "R(\"Hello World\")",
			expectedTokens: []itemType{itemIdent, itemLeftParen, itemString, itemRightParen},
		},
		"Number in relation symbol": {
			query:          "R(+51243)",
			expectedTokens: []itemType{itemIdent, itemLeftParen, itemInt, itemRightParen},
		},
		"Bool in relation symbol": {
			query:          "R(true)",
			expectedTokens: []itemType{itemIdent, itemLeftParen, itemBool, itemRightParen},
		},
		"Exist Quantor": {
			query:          "exists x in items: R(x)",
			expectedTokens: []itemType{itemExistQuant, itemIdent, itemIn, itemIdent, itemColon, itemIdent, itemLeftParen, itemIdent, itemRightParen},
		},
		"ForAll Quantor": {
			query:          "forall x in items: R(x)",
			expectedTokens: []itemType{itemAllQuant, itemIdent, itemIn, itemIdent, itemColon, itemIdent, itemLeftParen, itemIdent, itemRightParen},
		},
		"Multiple arguments": {
			query:          "R(abc, def, xyz)",
			expectedTokens: []itemType{itemIdent, itemLeftParen, itemIdent, itemComma, itemIdent, itemComma, itemIdent, itemRightParen},
		},
		"unnecessary but legal whitespace": {
			query:          "        R   (  x      )      ",
			expectedTokens: []itemType{itemIdent, itemLeftParen, itemIdent, itemRightParen},
		},
		"curly braces can be used as parens": {
			query:          "R{x}",
			expectedTokens: []itemType{itemIdent, itemLeftParen, itemIdent, itemRightParen},
		},
		"and can be used for &&": {
			query:          "true and false",
			expectedTokens: []itemType{itemBool, itemAnd, itemBool},
		},
		"or can be used for ||": {
			query:          "true or false",
			expectedTokens: []itemType{itemBool, itemOr, itemBool},
		},
		"impl can be used for ->": {
			query:          "true impl false",
			expectedTokens: []itemType{itemBool, itemImpl, itemBool},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lexer := lex(tc.query)
			tokens := make([]itemType, 0, len(tc.expectedTokens))
			for {
				nextToken := lexer.nextItem()
				assert.NotEqual(t, nextToken.typ, itemErr, "%s", nextToken.val)
				if nextToken.typ == eof {
					break
				}
				tokens = append(tokens, nextToken.typ)
			}
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}

func Test_lexLoadsCorrectIdentifierNameIntoValue(t *testing.T) {
	lexer := lex("R(myIdentifier)")
	lexer.nextItem()
	lexer.nextItem()
	id := lexer.nextItem()
	assert.Equal(t, itemIdent, id.typ)
	assert.Equal(t, "myIdentifier", id.val)
	assert.Equal(t, 2, id.pos)
}

func Test_lexOnInvalidQueries(t *testing.T) {
	testCases := map[string]struct {
		query string
	}{
		"Missing closing paren": {
			query: "R(x, K(y)",
		},
		"Too many closing paren": {
			query: "R((x, y)",
		},
		"Illegal identifier name": {
			query: "a/b",
		},
		"Illegal integer constant": {
			query: "0123f",
		},
		"Unclosed string literal": {
			query: "R(\"Hello World)",
		},
		"Illegal and junction": {
			query: "R(x) & P(x)",
		},
		"Illegal or junction": {
			query: "R(x) | P(x)",
		},
		"Illegal impl unction": {
			query: "R(x) - P(x)",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lexer := lex(tc.query)
			var err item
			for {
				nextToken := lexer.nextItem()
				if nextToken.typ == eof || nextToken.typ == itemErr {
					err = nextToken
					break
				}
			}
			assert.Equal(t, itemErr, err.typ)
		})
	}
}
