package query

import (
	"fmt"
	"strconv"

	"github.com/Fabian-G/todotxt/todotxt"
)

type varMap map[string]*todotxt.Item

type node interface {
	eval(todotxt.List, varMap) bool
	fmt.Stringer
}

type allQuant struct {
	boundId string
	child   node
}

func (a *allQuant) eval(l todotxt.List, alpha varMap) bool {
	prevValue := alpha[a.boundId]
	for _, item := range l {
		alpha[a.boundId] = item
		if !a.child.eval(l, alpha) {
			return false
		}
	}
	alpha[a.boundId] = prevValue
	return true
}

func (a *allQuant) String() string {
	return fmt.Sprintf("(forall %s: %s)", a.boundId, a.child.String())
}

type existQuant struct {
	boundId string
	child   node
}

func (e *existQuant) eval(l todotxt.List, alpha varMap) bool {
	prevValue := alpha[e.boundId]
	for _, item := range l {
		alpha[e.boundId] = item
		if e.child.eval(l, alpha) {
			return true
		}
	}
	alpha[e.boundId] = prevValue
	return false
}

func (e *existQuant) String() string {
	return fmt.Sprintf("(exists %s: %s)", e.boundId, e.child.String())
}

type impl struct {
	leftChild  node
	rightChild node
}

func (i *impl) eval(l todotxt.List, alpha varMap) bool {
	return !i.leftChild.eval(l, alpha) || i.rightChild.eval(l, alpha)
}

func (i *impl) String() string {
	return fmt.Sprintf("(%s -> %s)", i.leftChild.String(), i.rightChild.String())
}

type boolConst struct {
	val string
}

func (b *boolConst) eval(l todotxt.List, alpha varMap) bool {
	bo, _ := strconv.ParseBool(b.val)
	return bo
}

func (b *boolConst) String() string {
	return fmt.Sprintf("%s", b.val)
}
