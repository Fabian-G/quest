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
	list       *todotxt.List
	tags       map[string]qselect.DType
	unkownTags bool
	nowFunc    func() time.Time
}

func NewTagExpansion(list *todotxt.List, unknownTags bool, tags map[string]qselect.DType) todotxt.Hook {
	return &TagExpansion{
		list:       list,
		tags:       tags,
		unkownTags: unknownTags,
	}
}

func NewTagExpansionWithNowFunc(list *todotxt.List, unknownTags bool, tags map[string]qselect.DType, now func() time.Time) todotxt.Hook {
	return &TagExpansion{
		list:       list,
		tags:       tags,
		unkownTags: unknownTags,
		nowFunc:    now,
	}
}

func (t TagExpansion) OnMod(event todotxt.ModEvent) error {
	if event.Current == nil {
		return nil // We don't care about removals
	}
	for itemTag := range event.Current.Tags() {
		if typ, ok := t.tags[itemTag]; ok {
			t.expandTag(typ, itemTag, event.Current)
		}
	}
	return t.validateItem(event.Current)
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
	return t.validateItem(event.Item)
}

func (t TagExpansion) validateItem(item *todotxt.Item) error {
	var validationErrors []error
	for itemTag, tagValues := range item.Tags() {
		if typ, ok := t.tags[itemTag]; ok {
			validationErrors = append(validationErrors, validateTag(typ, itemTag, tagValues))
		} else if !t.unkownTags {
			validationErrors = append(validationErrors, fmt.Errorf("encountered unknown tag %s", itemTag))
		}
	}
	return errors.Join(validationErrors...)
}

func validateTag(typ qselect.DType, tag string, values []string) error {
	var validationErrors []error
	for _, v := range values {
		switch typ {
		case qselect.QInt:
			_, err := strconv.Atoi(v)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Errorf("tag \"%s\" of item violates int constraint: %w", tag, err))
			}
		case qselect.QDate:
			_, err := time.Parse(time.DateOnly, v)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Errorf("tag \"%s\" of item violates date constraint: %w", tag, err))
			}
		case qselect.QDuration:
			_, err := qduration.Parse(v)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Errorf("tag \"%s\" of item violates duration constraint: %w", tag, err))
			}
		case qselect.QBool:
			_, err := strconv.ParseBool(v)
			if err != nil {
				validationErrors = append(validationErrors, fmt.Errorf("tag \"%s\" of item violates bool constraint: %w", tag, err))
			}
		case qselect.QString:
			// String can be anythin
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
