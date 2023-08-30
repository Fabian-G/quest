package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Fabian-G/todotxt/todotxt"
)

type varMap map[string]*todotxt.Item

type node interface {
	eval(todotxt.List, varMap) any
	fmt.Stringer
}

type allQuant struct {
	boundId string
	child   node
}

func (a *allQuant) eval(l todotxt.List, alpha varMap) any {
	prevValue := alpha[a.boundId]
	for _, item := range l {
		alpha[a.boundId] = item
		if !a.child.eval(l, alpha).(bool) {
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

func (e *existQuant) eval(l todotxt.List, alpha varMap) any {
	prevValue := alpha[e.boundId]
	for _, item := range l {
		alpha[e.boundId] = item
		if e.child.eval(l, alpha).(bool) {
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

func (i *impl) eval(l todotxt.List, alpha varMap) any {
	return !i.leftChild.eval(l, alpha).(bool) || i.rightChild.eval(l, alpha).(bool)
}

func (i *impl) String() string {
	return fmt.Sprintf("(%s -> %s)", i.leftChild.String(), i.rightChild.String())
}

type and struct {
	leftChild  node
	rightChild node
}

func (a *and) eval(l todotxt.List, alpha varMap) any {
	return a.leftChild.eval(l, alpha).(bool) && a.rightChild.eval(l, alpha).(bool)
}

func (a *and) String() string {
	return fmt.Sprintf("(%s && %s)", a.leftChild.String(), a.rightChild.String())
}

type or struct {
	leftChild  node
	rightChild node
}

func (o *or) eval(l todotxt.List, alpha varMap) any {
	return o.leftChild.eval(l, alpha).(bool) || o.rightChild.eval(l, alpha).(bool)
}

func (o *or) String() string {
	return fmt.Sprintf("(%s || %s)", o.leftChild.String(), o.rightChild.String())
}

type not struct {
	child node
}

func (n *not) eval(l todotxt.List, alpha varMap) any {
	return !n.child.eval(l, alpha).(bool)
}

func (n *not) String() string {
	return fmt.Sprintf("!%s", n.child)
}

type stringConst struct {
	val string
}

func (s *stringConst) eval(l todotxt.List, alpha varMap) any {
	return s.val[1 : len(s.val)-1]
}

func (s *stringConst) String() string {
	return s.val
}

type intConst struct {
	val string
}

func (i *intConst) eval(l todotxt.List, alpha varMap) any {
	n, _ := strconv.Atoi(i.val)
	return n
}

func (i *intConst) String() string {
	return i.val
}

type boolConst struct {
	val string
}

func (b *boolConst) eval(l todotxt.List, alpha varMap) any {
	bo, _ := strconv.ParseBool(b.val)
	return bo
}

func (b *boolConst) String() string {
	return b.val
}

type identifier struct {
	name string
}

func (i *identifier) eval(l todotxt.List, alpha varMap) any {
	return alpha[i.name]
}

func (i *identifier) String() string {
	return i.name
}

type call struct {
	name string
	args node
}

func (c *call) eval(l todotxt.List, alpha varMap) any {
	fn := functions[c.name]
	return fn(c.args.eval(l, alpha).([]any))
}

func (i *call) String() string {
	return fmt.Sprintf("%s%s", i.name, i.args.String())
}

type args struct {
	children []node
}

func (a *args) eval(l todotxt.List, alpha varMap) any {
	result := make([]any, 0, len(a.children))
	for _, c := range a.children {
		result = append(result, c.eval(l, alpha))
	}
	return result
}

func (a *args) String() string {
	argsList := strings.Builder{}
	for _, arg := range a.children {
		argsList.WriteString(arg.String())
		argsList.WriteString(", ")
	}
	argListString := argsList.String()
	return fmt.Sprintf("(%s)", argListString[:max(0, len(argListString)-2)])
}
