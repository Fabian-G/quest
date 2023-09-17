package query

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

type DType string

const (
	QError       DType = "error"
	QInt         DType = "int"
	QDate        DType = "date"
	QDuration    DType = "duration"
	QString      DType = "string"
	QStringSlice DType = "[]string"
	QBool        DType = "bool"
	QItem        DType = "item"
	QItemSlice   DType = "[]item"
)

var AllDTypes = []DType{QInt, QDate, QDuration, QString, QStringSlice, QBool, QItem, QItemSlice}

func (d DType) isSliceType() bool {
	return slices.Contains([]DType{QStringSlice, QItemSlice}, d)
}

func (d DType) sliceTypeToItemType() DType {
	switch d {
	case QItemSlice:
		return QItem
	case QStringSlice:
		return QString
	}
	return QError
}

func toAnySlice[S ~[]E, E any](s S) []any {
	r := make([]any, len(s))
	for i, e := range s {
		r[i] = e
	}
	return r
}

type varMap map[string]any
type idSet map[string]DType

type node interface {
	validate(idSet) (DType, error)
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
	return fmt.Sprintf("(forall %s in %s: %s)", a.boundId, a.collection.String(), a.child.String())
}

func (a *allQuant) validate(knownIds idSet) (DType, error) {
	collectionType, err := a.collection.validate(knownIds)
	if err != nil {
		return QError, err
	}
	if !collectionType.isSliceType() {
		return QError, fmt.Errorf("can not use non slice type %s as collection in quantifier", collectionType)
	}
	if _, ok := knownIds[a.boundId]; !ok {
		knownIds[a.boundId] = collectionType.sliceTypeToItemType()
		defer delete(knownIds, a.boundId)
	}
	childType, err := a.child.validate(knownIds)
	if err != nil {
		return QError, err
	}
	if childType != QBool {
		return QError, fmt.Errorf("can not apply all quantor on expression of type %s", childType)
	}
	return QBool, nil
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
	return fmt.Sprintf("(exists %s in %s: %s)", e.boundId, e.collection.String(), e.child.String())
}

func (e *existQuant) validate(knownIds idSet) (DType, error) {
	collectionType, err := e.collection.validate(knownIds)
	if err != nil {
		return QError, err
	}
	if !collectionType.isSliceType() {
		return QError, fmt.Errorf("can not use non slice type %s as collection in quantifier", collectionType)
	}
	if _, ok := knownIds[e.boundId]; !ok {
		knownIds[e.boundId] = collectionType.sliceTypeToItemType()
		defer delete(knownIds, e.boundId)
	}
	childType, err := e.child.validate(knownIds)
	if err != nil {
		return QError, err
	}
	if childType != QBool {
		return QError, fmt.Errorf("can not apply exists quantor on expression of type %s", childType)
	}
	return QBool, nil
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

func (i *impl) validate(knownIds idSet) (DType, error) {
	c1, err := i.leftChild.validate(knownIds)
	if err != nil {
		return QError, err
	}
	c2, err := i.rightChild.validate(knownIds)
	if err != nil {
		return QError, err
	}
	if c1 != QBool || c2 != QBool {
		return QError, fmt.Errorf("can not apply implication on (%s, %s)", c1, c2)
	}
	return QBool, nil
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

func (a *and) validate(knownIds idSet) (DType, error) {
	c1, err := a.leftChild.validate(knownIds)
	if err != nil {
		return QError, err
	}
	c2, err := a.rightChild.validate(knownIds)
	if err != nil {
		return QError, err
	}
	if c1 != QBool || c2 != QBool {
		return QError, fmt.Errorf("can not apply conjunction on (%s, %s)", c1, c2)
	}
	return QBool, nil
}

type comparison struct {
	comparator itemType
	lType      DType
	rType      DType
	leftChild  node
	rightChild node
}

func (e *comparison) eval(alpha varMap) any {
	left, right := e.leftChild.eval(alpha), e.rightChild.eval(alpha)
	lType := e.lType
	rType := e.rType
	return compare(e.comparator, lType, rType, left, right)
}

func compare(op itemType, t1 DType, t2 DType, left, right any) bool {
	switch t1 {
	case QString:
		v1 := left.(string)
		v2 := right.(string)
		return compareComparable(op, v1, v2)
	case QItem:
		return left == right // Only eq comparison allowed. Wrong usage is already caught in validaton
	case QDate:
		v1 := left.(time.Time)
		v2 := right.(time.Time)
		return compareDates(op, v1, v2)
	case QDuration:
		v1 := left.(time.Duration)
		v2 := right.(time.Duration)
		return compareComparable(op, v1, v2)
	case QInt:
		v1 := left.(int)
		v2 := right.(int)
		return compareComparable(op, v1, v2)
	case QBool:
		v1 := left.(bool)
		v2 := right.(bool)
		var v1i, v2i int
		if v1 {
			v1i = 1
		}
		if v2 {
			v2i = 1
		}
		return compareComparable(op, v1i, v2i)
	default:
		panic(fmt.Errorf("comparing uncomparable types %s and %s", t1, t2))
	}
}

func compareComparable[E cmp.Ordered](op itemType, s1, s2 E) bool {
	switch op {
	case itemEq:
		return s1 == s2
	case itemLt:
		return s1 < s2
	case itemLeq:
		return s1 <= s2
	case itemGt:
		return s1 > s2
	case itemGeq:
		return s1 >= s2
	default:
		return false
	}
}

func compareDates(op itemType, d1, d2 time.Time) bool {
	switch op {
	case itemEq:
		return d1.Equal(d2)
	case itemLt:
		return d1.Before(d2)
	case itemLeq:
		return d1.Before(d2) || d1.Equal(d2)
	case itemGt:
		return d1.After(d2)
	case itemGeq:
		return d1.After(d2) || d1.Equal(d2)
	default:
		return false
	}
}
func (e *comparison) String() string {
	var opString string
	switch e.comparator {
	case itemEq:
		opString = "=="
	case itemLt:
		opString = "<"
	case itemLeq:
		opString = "<="
	case itemGt:
		opString = ">"
	case itemGeq:
		opString = ">="
	}
	return fmt.Sprintf("(%s %s %s)", e.leftChild.String(), opString, e.rightChild.String())
}

func (e *comparison) validate(knownIds idSet) (DType, error) {
	leftType, err := e.leftChild.validate(knownIds)
	if err != nil {
		return QError, err
	}
	rightType, err := e.rightChild.validate(knownIds)
	if err != nil {
		return QError, err
	}
	if leftType != rightType {
		return QError, fmt.Errorf("can not compare %s with %s", leftType, rightType)
	}
	if leftType.isSliceType() || rightType.isSliceType() {
		return QError, errors.New("comparing slice types is not allowed")
	}
	if leftType == QItem && e.comparator != itemEq {
		return QError, errors.New("items can only be compared using ==")
	}
	allowedTypes := []DType{QString, QItem, QDate, QDuration, QInt, QBool}
	if !slices.Contains(allowedTypes, leftType) || !slices.Contains(allowedTypes, rightType) {
		return QError, fmt.Errorf("can not compare %s with %s. Allowed types are: %v", leftType, rightType, allowedTypes)
	}
	e.lType = leftType
	e.rType = rightType
	return QBool, nil
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

func (o *or) validate(knownIds idSet) (DType, error) {
	c1, err := o.leftChild.validate(knownIds)
	if err != nil {
		return QError, err
	}
	c2, err := o.rightChild.validate(knownIds)
	if err != nil {
		return QError, err
	}
	if c1 != QBool || c2 != QBool {
		return QError, fmt.Errorf("can not apply disjunction on (%s, %s)", c1, c2)
	}
	return QBool, nil
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

func (n *not) validate(knownIds idSet) (DType, error) {
	c, err := n.child.validate(knownIds)
	if err != nil {
		return QError, err
	}
	if c != QBool {
		return QError, fmt.Errorf("can not apply not on %s", c)
	}
	return QBool, nil
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
func (s *stringConst) validate(knownIds idSet) (DType, error) {
	return QString, nil
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

func (i *intConst) validate(knownIds idSet) (DType, error) {
	if _, err := strconv.Atoi(i.val); err != nil {
		return QError, fmt.Errorf("could not parse integer constant: %s", i.val)
	}
	return QInt, nil
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

func (b *boolConst) validate(knownIds idSet) (DType, error) {
	if _, err := strconv.ParseBool(b.val); err != nil {
		return QError, fmt.Errorf("could not parse boolean constant: %s", b.val)
	}
	return QBool, nil
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

func (i *identifier) validate(knownIds idSet) (DType, error) {
	if _, ok := knownIds[i.name]; !ok {
		return QError, fmt.Errorf("unknown identifier: %s", i.name)
	}
	return knownIds[i.name], nil
}

type call struct {
	name        string
	fn          queryFunc
	args        *args
	ifBound     node
	passThrough bool // if true, all calls will get passed to ifBound. This field is set by validate()
}

func (c *call) eval(alpha varMap) any {
	if c.passThrough {
		return c.ifBound.eval(alpha)
	}
	return c.fn.call(alpha, c.args.eval(alpha).([]any))
}

func (c *call) String() string {
	if c.passThrough {
		return c.ifBound.String()
	}
	return fmt.Sprintf("%s%s", c.name, c.args.String())
}

func (c *call) validate(knownIds idSet) (DType, error) {
	if _, ok := knownIds[c.name]; ok && c.ifBound != nil {
		c.passThrough = true
		return c.ifBound.validate(knownIds)
	}
	argTypes, err := c.args.validateAll(knownIds)
	if err != nil {
		return QError, err
	}
	if c.fn.fn == nil {
		return QError, fmt.Errorf("unknown function with name %s", c.name)
	}
	err = c.fn.validate(argTypes)
	var missingItemError missingItemError
	if errors.As(err, &missingItemError) {
		it := identifier{name: "it"}
		c.args.children = slices.Insert[[]node, node](c.args.children, missingItemError.position, &it)
		return c.validate(knownIds)
	}
	return c.fn.resultType, err
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

func (a *args) validate(knownIds idSet) (DType, error) {
	return QError, errors.New("args must only occur in the context of a function call")
}

func (a *args) validateAll(knownIds idSet) ([]DType, error) {
	types := make([]DType, 0, len(a.children))
	for _, c := range a.children {
		t, err := c.validate(knownIds)
		if err != nil {
			return []DType{QError}, err
		}
		types = append(types, t)
	}
	return types, nil
}
