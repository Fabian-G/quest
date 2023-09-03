package todotxt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AssertItemEqual compares two items
// This is mainly needed, because reflect.DeppeEqual can't compare function field, but Item has one
func AssertItemEqual(t *testing.T, expected *Item, actual *Item) {
	if expected != nil {
		emitFuncBeforeExpected := expected.emitFunc
		expected.emitFunc = nil
		defer func() { expected.emitFunc = emitFuncBeforeExpected }()
	}
	if actual != nil {
		emitFuncBeforeActual := actual.emitFunc
		actual.emitFunc = nil
		defer func() { actual.emitFunc = emitFuncBeforeActual }()
	}
	assert.Equal(t, expected, actual)
}

func AssertListEqual(t *testing.T, expected *List, actual *List) {
	assert.Equal(t, expected.Len(), actual.Len())
	for i, task := range expected.Tasks() {
		AssertItemEqual(t, task, actual.Get(i))
	}
}
