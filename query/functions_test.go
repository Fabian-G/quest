package query

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryFunc_validate(t *testing.T) {
	testcases := map[string]struct {
		fn      queryFunc
		actual  []dType
		allowed bool
	}{
		"one missing it": {
			fn: queryFunc{
				argTypes:         []dType{qBool, qItem, qBool},
				trailingOptional: false,
				injectIt:         true,
			},
			actual:  []dType{qBool, qBool},
			allowed: true,
		},
		"Exact args": {
			fn: queryFunc{
				argTypes:         []dType{qBool, qItem, qBool},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []dType{qBool, qItem, qBool},
			allowed: true,
		},
		"Inserting it leads to too many args": {
			fn: queryFunc{
				argTypes:         []dType{qItem, qItem, qItem},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []dType{qItem, qItem, qBool},
			allowed: false,
		},
		"It on zero args can be inserted": {
			fn: queryFunc{
				argTypes:         []dType{qItem},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []dType{},
			allowed: true,
		},
		"Trailing Optional can be omitted": {
			fn: queryFunc{
				argTypes:         []dType{qBool},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []dType{},
			allowed: true,
		},
		"Trailing Optional can be omitted (with many args)": {
			fn: queryFunc{
				argTypes:         []dType{qBool, qBool, qBool},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []dType{qBool, qBool},
			allowed: true,
		},
		"Only one trailing arg can be omitted": {
			fn: queryFunc{
				argTypes:         []dType{qBool, qBool, qBool},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []dType{qBool},
			allowed: false,
		},
		"Injecting it and trailing optional can happen at the same time": {
			fn: queryFunc{
				argTypes:         []dType{qBool, qItem, qBool, qBool},
				trailingOptional: true,
				injectIt:         true,
			},
			actual:  []dType{qBool, qBool},
			allowed: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			err := tc.fn.validate(tc.actual)
			var missingItemError missingItemError
			if errors.As(err, &missingItemError) {
				tc.actual = slices.Insert(tc.actual, missingItemError.position, qItem)
			}
			err = tc.fn.validate(tc.actual)
			if tc.allowed {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
