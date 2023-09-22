package hook

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Fabian-G/quest/qduration"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
)

type TagExpansion struct {
	list    *todotxt.List
	tags    map[string]qselect.DType
	nowFunc func() time.Time
}

func NewTagExpansion(list *todotxt.List, tags map[string]qselect.DType) todotxt.Hook {
	return &TagExpansion{
		list: list,
		tags: tags,
	}
}

func NewTagExpansionWithNowFunc(list *todotxt.List, tags map[string]qselect.DType, now func() time.Time) todotxt.Hook {
	return &TagExpansion{
		list:    list,
		tags:    tags,
		nowFunc: now,
	}
}

func (t TagExpansion) OnMod(event todotxt.ModEvent) error {
	var validationErrors []error
	for itemTag := range event.Current.Tags() {
		if typ, ok := t.tags[itemTag]; ok {
			t.expandTag(typ, itemTag, event.Current)
			validationErrors = append(validationErrors, validate(typ, event.Current.Tags()[itemTag]))
		}
	}
	return errors.Join(validationErrors...)
}

func (t TagExpansion) expandTag(typ qselect.DType, tag string, i *todotxt.Item) {
	tagValue := i.Tags()[tag][0]
	switch typ {
	case qselect.QDate:
		i.SetTag(tag, t.expandDate(tagValue))
	case qselect.QInt:
		i.SetTag(tag, t.expandInt(i, tag, tagValue))
	}
}

func (t TagExpansion) expandDate(v string) string {
	if _, err := time.Parse(time.DateOnly, v); err == nil {
		// This is already a proper date
		return v
	}
	var base time.Time
	var remainingValue string = v
	switch {
	case strings.HasPrefix(v, "today"):
		remainingValue = strings.TrimPrefix(v, "today")
		fallthrough
	default:
		base = t.now()
	}
	if len(strings.TrimSpace(remainingValue)) == 0 {
		return base.Format(time.DateOnly)
	}
	if duration, err := qduration.Parse(remainingValue); err == nil {
		// This is a duration
		return duration.AddTo(base).Format(time.DateOnly)
	}
	// Unknown expansion. Let validation handle that
	return v
}

func (t TagExpansion) expandInt(i *todotxt.Item, tag string, value string) string {
	if _, err := strconv.Atoi(value); err == nil {
		// This is already a proper integer
		return value
	}

	var base int
	var remainingValue string
	switch {
	case strings.HasPrefix(value, "max"):
		base = t.globalMaxValue(tag)
		remainingValue = strings.TrimPrefix(value, "max")
	case strings.HasPrefix(value, "min"):
		base = t.globalMinValue(tag)
		remainingValue = strings.TrimPrefix(value, "min")
	case strings.HasPrefix(value, "pmax"):
		base = t.projectMaxValue(i.Projects(), tag)
		remainingValue = strings.TrimPrefix(value, "pmax")
	case strings.HasPrefix(value, "pmin"):
		base = t.projectMinValue(i.Projects(), tag)
		remainingValue = strings.TrimPrefix(value, "pmin")
	}
	if len(strings.TrimSpace(remainingValue)) == 0 {
		return strconv.Itoa(base)
	}
	offset, err := strconv.Atoi(remainingValue)
	if err != nil {
		return value
	}
	return strconv.Itoa(base + offset)
}

func (t TagExpansion) globalMaxValue(tag string) int {
	var maximum int = math.MinInt
	var foundOne bool
	for _, i := range t.list.Tasks() {
		for _, v := range i.Tags()[tag] {
			if i, err := strconv.Atoi(v); err == nil {
				foundOne = true
				maximum = max(i, maximum)
			}
		}
	}
	if !foundOne {
		return 0
	}
	return maximum
}

func (t TagExpansion) globalMinValue(tag string) int {
	var minimum int = math.MaxInt
	var foundOne bool
	for _, i := range t.list.Tasks() {
		for _, v := range i.Tags()[tag] {
			if i, err := strconv.Atoi(v); err == nil {
				foundOne = true
				minimum = min(i, minimum)
			}
		}
	}
	if !foundOne {
		return 0
	}
	return minimum
}

func (t TagExpansion) projectMaxValue(projects []todotxt.Project, tag string) int {
	var maximum int = math.MinInt
	var foundOne bool
	for _, i := range t.list.Tasks() {
		if !slices.ContainsFunc(i.Projects(), func(p todotxt.Project) bool {
			return slices.Contains(projects, p)
		}) {
			continue
		}
		for _, v := range i.Tags()[tag] {
			if i, err := strconv.Atoi(v); err == nil {
				foundOne = true
				maximum = max(i, maximum)
			}
		}
	}
	if !foundOne {
		return 0
	}
	return maximum
}

func (t TagExpansion) projectMinValue(projects []todotxt.Project, tag string) int {
	var minimum int = math.MaxInt
	var foundOne bool
	for _, i := range t.list.Tasks() {
		if !slices.ContainsFunc(i.Projects(), func(p todotxt.Project) bool {
			return slices.Contains(projects, p)
		}) {
			continue
		}
		for _, v := range i.Tags()[tag] {
			if i, err := strconv.Atoi(v); err == nil {
				foundOne = true
				minimum = min(i, minimum)
			}
		}
	}
	if !foundOne {
		return 0
	}
	return minimum
}

func (t TagExpansion) OnValidate(event todotxt.ValidationEvent) error {
	var validationErrors []error
	for itemTag, tagValues := range event.Item.Tags() {
		if typ, ok := t.tags[itemTag]; ok {
			validationErrors = append(validationErrors, validate(typ, tagValues))
		}
	}
	return errors.Join(validationErrors...)
}

func validate(typ qselect.DType, values []string) error {
	var validationErrors []error
	for _, v := range values {
		switch typ {
		case qselect.QInt:
			_, err := strconv.Atoi(v)
			validationErrors = append(validationErrors, err)
		case qselect.QDate:
			_, err := time.Parse(time.DateOnly, v)
			validationErrors = append(validationErrors, err)
		case qselect.QDuration:
			_, err := qduration.Parse(v)
			validationErrors = append(validationErrors, err)
		case qselect.QBool:
			_, err := strconv.ParseBool(v)
			validationErrors = append(validationErrors, err)
		default:
			return fmt.Errorf("validation for type %s is not supported", typ)
		}
	}
	return errors.Join(validationErrors...)
}

func (t TagExpansion) now() time.Time {
	if t.nowFunc != nil {
		return t.nowFunc()
	}
	return time.Now()
}
