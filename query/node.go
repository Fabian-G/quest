package query

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Fabian-G/todotxt/todotxt"
)

type dType string

const (
	qError       dType = "error"
	qInt         dType = "int"
	qString      dType = "string"
	qStringSlice dType = "[]string"
	qBool        dType = "bool"
	qItem        dType = "item"
)

type varMap map[string]*todotxt.Item

type node interface {
	validate() (dType, error)
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

func (a *allQuant) validate() (dType, error) {
	childType, err := a.child.validate()
	if err != nil {
		return qError, err
	}
	if childType != qBool {
		return qError, fmt.Errorf("can not apply all quantor on expression of type %s", childType)
	}
	return qBool, nil
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

func (e *existQuant) validate() (dType, error) {
	childType, err := e.child.validate()
	if err != nil {
		return qError, err
	}
	if childType != qBool {
		return qError, fmt.Errorf("can not apply exists quantor on expression of type %s", childType)
	}
	return qBool, nil
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

func (i *impl) validate() (dType, error) {
	c1, err := i.leftChild.validate()
	if err != nil {
		return qError, err
	}
	c2, err := i.rightChild.validate()
	if err != nil {
		return qError, err
	}
	if c1 != qBool || c2 != qBool {
		return qError, fmt.Errorf("can not apply implication on (%s, %s)", c1, c2)
	}
	return qBool, nil
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

func (a *and) validate() (dType, error) {
	c1, err := a.leftChild.validate()
	if err != nil {
		return qError, err
	}
	c2, err := a.rightChild.validate()
	if err != nil {
		return qError, err
	}
	if c1 != qBool || c2 != qBool {
		return qError, fmt.Errorf("can not apply conjunction on (%s, %s)", c1, c2)
	}
	return qBool, nil
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

func (o *or) validate() (dType, error) {
	c1, err := o.leftChild.validate()
	if err != nil {
		return qError, err
	}
	c2, err := o.rightChild.validate()
	if err != nil {
		return qError, err
	}
	if c1 != qBool || c2 != qBool {
		return qError, fmt.Errorf("can not apply disjunction on (%s, %s)", c1, c2)
	}
	return qBool, nil
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

func (n *not) validate() (dType, error) {
	c, err := n.child.validate()
	if err != nil {
		return qError, err
	}
	if c != qBool {
		return qError, fmt.Errorf("can not apply not on %s", c)
	}
	return qBool, nil
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
func (s *stringConst) validate() (dType, error) {
	return qString, nil
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

func (i *intConst) validate() (dType, error) {
	if _, err := strconv.Atoi(i.val); err != nil {
		return qError, fmt.Errorf("could not parse integer constant: %s", i.val)
	}
	return qString, nil
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

func (b *boolConst) validate() (dType, error) {
	if _, err := strconv.ParseBool(b.val); err != nil {
		return qError, fmt.Errorf("could not parse boolean constant: %s", b.val)
	}
	return qBool, nil
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

func (i *identifier) validate() (dType, error) {
	return qItem, nil
}

type call struct {
	name string
	args *args
}

func (c *call) eval(l todotxt.List, alpha varMap) any {
	fn := functions[c.name]
	return fn(c.args.eval(l, alpha).([]any))
}

func (c *call) String() string {
	return fmt.Sprintf("%s%s", c.name, c.args.String())
}

func (c *call) validate() (dType, error) {
	argTypes, err := c.args.validateAll()
	if err != nil {
		return qError, err
	}
	return funcType(c.name, argTypes)
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

func (a *args) validate() (dType, error) {
	return qError, errors.New("args must only occur in the context of a function call")
}

func (a *args) validateAll() ([]dType, error) {
	types := make([]dType, 0, len(a.children))
	for _, c := range a.children {
		t, err := c.validate()
		if err != nil {
			return []dType{qError}, err
		}
		types = append(types, t)
	}
	return types, nil
}
