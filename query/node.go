package query

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type dType string

const (
	qError       dType = "error"
	qInt         dType = "int"
	qString      dType = "string"
	qStringSlice dType = "[]string"
	qBool        dType = "bool"
	qItem        dType = "item"
	qItemSlice   dType = "[]item"
)

func (d dType) isSliceType() bool {
	return slices.Contains([]dType{qStringSlice, qItemSlice}, d)
}

func (d dType) sliceTypeToItemType() dType {
	switch d {
	case qItemSlice:
		return qItem
	case qStringSlice:
		return qString
	}
	return qError
}

func toAnySlice[S ~[]E, E any](s S) []any {
	r := make([]any, len(s))
	for i, e := range s {
		r[i] = e
	}
	return r
}

type varMap map[string]any
type idSet map[string]dType

type node interface {
	validate(idSet) (dType, error)
	eval(varMap) any
	fmt.Stringer
}

type allQuant struct {
	boundId    string
	child      node
	collection node
}

func (a *allQuant) eval(alpha varMap) any {
	prevValue := alpha[a.boundId]
	defer func() { alpha[a.boundId] = prevValue }()
	for _, item := range a.collection.eval(alpha).([]any) {
		alpha[a.boundId] = item
		if !a.child.eval(alpha).(bool) {
			return false
		}
	}
	return true
}

func (a *allQuant) String() string {
	return fmt.Sprintf("(forall %s in (%s): %s)", a.boundId, a.collection.String(), a.child.String())
}

func (a *allQuant) validate(knownIds idSet) (dType, error) {
	collectionType, err := a.collection.validate(knownIds)
	if err != nil {
		return qError, err
	}
	if !collectionType.isSliceType() {
		return qError, fmt.Errorf("can not use non slice type %s as collection in quantifier", collectionType)
	}
	if _, ok := knownIds[a.boundId]; !ok {
		knownIds[a.boundId] = collectionType.sliceTypeToItemType()
		defer delete(knownIds, a.boundId)
	}
	childType, err := a.child.validate(knownIds)
	if err != nil {
		return qError, err
	}
	if childType != qBool {
		return qError, fmt.Errorf("can not apply all quantor on expression of type %s", childType)
	}
	return qBool, nil
}

type existQuant struct {
	boundId    string
	child      node
	collection node
}

func (e *existQuant) eval(alpha varMap) any {
	prevValue := alpha[e.boundId]
	defer func() { alpha[e.boundId] = prevValue }()
	for _, item := range e.collection.eval(alpha).([]any) {
		alpha[e.boundId] = item
		if e.child.eval(alpha).(bool) {
			return true
		}
	}
	return false
}

func (e *existQuant) String() string {
	return fmt.Sprintf("(exists %s in (%s): %s)", e.boundId, e.collection.String(), e.child.String())
}

func (e *existQuant) validate(knownIds idSet) (dType, error) {
	collectionType, err := e.collection.validate(knownIds)
	if err != nil {
		return qError, err
	}
	if !collectionType.isSliceType() {
		return qError, fmt.Errorf("can not use non slice type %s as collection in quantifier", collectionType)
	}
	if _, ok := knownIds[e.boundId]; !ok {
		knownIds[e.boundId] = collectionType.sliceTypeToItemType()
		defer delete(knownIds, e.boundId)
	}
	childType, err := e.child.validate(knownIds)
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

func (i *impl) eval(alpha varMap) any {
	return !i.leftChild.eval(alpha).(bool) || i.rightChild.eval(alpha).(bool)
}

func (i *impl) String() string {
	return fmt.Sprintf("(%s -> %s)", i.leftChild.String(), i.rightChild.String())
}

func (i *impl) validate(knownIds idSet) (dType, error) {
	c1, err := i.leftChild.validate(knownIds)
	if err != nil {
		return qError, err
	}
	c2, err := i.rightChild.validate(knownIds)
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

func (a *and) eval(alpha varMap) any {
	return a.leftChild.eval(alpha).(bool) && a.rightChild.eval(alpha).(bool)
}

func (a *and) String() string {
	return fmt.Sprintf("(%s && %s)", a.leftChild.String(), a.rightChild.String())
}

func (a *and) validate(knownIds idSet) (dType, error) {
	c1, err := a.leftChild.validate(knownIds)
	if err != nil {
		return qError, err
	}
	c2, err := a.rightChild.validate(knownIds)
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

func (o *or) eval(alpha varMap) any {
	return o.leftChild.eval(alpha).(bool) || o.rightChild.eval(alpha).(bool)
}

func (o *or) String() string {
	return fmt.Sprintf("(%s || %s)", o.leftChild.String(), o.rightChild.String())
}

func (o *or) validate(knownIds idSet) (dType, error) {
	c1, err := o.leftChild.validate(knownIds)
	if err != nil {
		return qError, err
	}
	c2, err := o.rightChild.validate(knownIds)
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

func (n *not) eval(alpha varMap) any {
	return !n.child.eval(alpha).(bool)
}

func (n *not) String() string {
	return fmt.Sprintf("!%s", n.child)
}

func (n *not) validate(knownIds idSet) (dType, error) {
	c, err := n.child.validate(knownIds)
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

func (s *stringConst) eval(alpha varMap) any {
	return s.val[1 : len(s.val)-1]
}

func (s *stringConst) String() string {
	return s.val

}
func (s *stringConst) validate(knownIds idSet) (dType, error) {
	return qString, nil
}

type intConst struct {
	val string
}

func (i *intConst) eval(alpha varMap) any {
	n, _ := strconv.Atoi(i.val)
	return n
}

func (i *intConst) String() string {
	return i.val
}

func (i *intConst) validate(knownIds idSet) (dType, error) {
	if _, err := strconv.Atoi(i.val); err != nil {
		return qError, fmt.Errorf("could not parse integer constant: %s", i.val)
	}
	return qString, nil
}

type boolConst struct {
	val string
}

func (b *boolConst) eval(alpha varMap) any {
	bo, _ := strconv.ParseBool(b.val)
	return bo
}

func (b *boolConst) String() string {
	return b.val
}

func (b *boolConst) validate(knownIds idSet) (dType, error) {
	if _, err := strconv.ParseBool(b.val); err != nil {
		return qError, fmt.Errorf("could not parse boolean constant: %s", b.val)
	}
	return qBool, nil
}

type identifier struct {
	name string
}

func (i *identifier) eval(alpha varMap) any {
	return alpha[i.name]
}

func (i *identifier) String() string {
	return i.name
}

func (i *identifier) validate(knownIds idSet) (dType, error) {
	if _, ok := knownIds[i.name]; !ok {
		return qError, fmt.Errorf("unknown identifier: %s", i.name)
	}
	return knownIds[i.name], nil
}

type call struct {
	name        string
	args        *args
	ifBound     node
	passThrough bool // if true, all calls will get passed to ifBound. This field is set by validate()
}

func (c *call) eval(alpha varMap) any {
	if c.passThrough {
		return c.ifBound.eval(alpha)
	}
	fn := functions[c.name]
	return fn.call(c.args.eval(alpha).([]any))
}

func (c *call) String() string {
	if c.passThrough {
		return c.ifBound.String()
	}
	return fmt.Sprintf("%s%s", c.name, c.args.String())
}

func (c *call) validate(knownIds idSet) (dType, error) {
	if _, ok := knownIds[c.name]; ok && c.ifBound != nil {
		c.passThrough = true
		return c.ifBound.validate(knownIds)
	}
	argTypes, err := c.args.validateAll(knownIds)
	if err != nil {
		return qError, err
	}
	var fn queryFunc
	var ok bool
	if fn, ok = functions[c.name]; !ok {
		return qError, fmt.Errorf("unknown function with name %s", c.name)
	}
	err = fn.validate(argTypes)
	var missingItemError missingItemError
	if errors.As(err, &missingItemError) {
		it := identifier{name: "it"}
		c.args.children = slices.Insert[[]node, node](c.args.children, missingItemError.position, &it)
		return c.validate(knownIds)
	}
	return fn.resultType, err
}

type args struct {
	children []node
}

func (a *args) eval(alpha varMap) any {
	result := make([]any, 0, len(a.children))
	for _, c := range a.children {
		result = append(result, c.eval(alpha))
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

func (a *args) validate(knownIds idSet) (dType, error) {
	return qError, errors.New("args must only occur in the context of a function call")
}

func (a *args) validateAll(knownIds idSet) ([]dType, error) {
	types := make([]dType, 0, len(a.children))
	for _, c := range a.children {
		t, err := c.validate(knownIds)
		if err != nil {
			return []dType{qError}, err
		}
		types = append(types, t)
	}
	return types, nil
}
