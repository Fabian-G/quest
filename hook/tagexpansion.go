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

var weekdayLookup = map[string]time.Weekday{
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
	"sunday":    time.Sunday,
	"mon":       time.Monday,
	"tue":       time.Tuesday,
	"wed":       time.Wednesday,
	"thu":       time.Thursday,
	"fri":       time.Friday,
	"sat":       time.Saturday,
	"sun":       time.Sunday,
}

type TagExpansion struct {
	tags       map[string]qselect.DType
	unkownTags bool
	nowFunc    func() time.Time
}

func NewTagExpansion(unknownTags bool, tags map[string]qselect.DType) todotxt.Hook {
	return &TagExpansion{
		tags:       tags,
		unkownTags: unknownTags,
	}
}

func NewTagExpansionWithNowFunc(unknownTags bool, tags map[string]qselect.DType, now func() time.Time) todotxt.Hook {
	return &TagExpansion{
		tags:       tags,
		unkownTags: unknownTags,
		nowFunc:    now,
	}
}

func (t TagExpansion) OnMod(list *todotxt.List, event todotxt.ModEvent) error {
	if event.Current == nil {
		return nil // We don't care about removals
	}
	for itemTag := range event.Current.Tags() {
		if typ, ok := t.tags[itemTag]; ok {
			if err := t.expandTag(list, typ, itemTag, event.Current); err != nil {
				return err
			}
		}
	}
	return t.validateItem(event.Current)
}

func (t TagExpansion) expandTag(list *todotxt.List, typ qselect.DType, tag string, i *todotxt.Item) error {
	tagValue := i.Tags()[tag][0]
	switch typ {
	case qselect.QDate:
		return i.SetTag(tag, t.expandDate(tagValue))
	case qselect.QInt:
		return i.SetTag(tag, t.expandInt(list, i, tag, tagValue))
	}
	return nil
}

func (t TagExpansion) expandDate(v string) string {
	if _, err := time.Parse(time.DateOnly, v); err == nil {
		// This is already a proper date
		return v
	}
	var base time.Time
	var remainingValue string = strings.ToLower(v)
	possibleBase := t.guessBase(remainingValue)
	switch {
	case t.isWeekday(possibleBase):
		remainingValue = strings.TrimPrefix(remainingValue, possibleBase)
		base = t.findNext(weekdayLookup[possibleBase])
	case possibleBase == "tomorrow":
		base = t.today().Add(24 * time.Hour)
		remainingValue = strings.TrimPrefix(remainingValue, possibleBase)
	case possibleBase == "yesterday":
		base = t.today().Add(-24 * time.Hour)
		remainingValue = strings.TrimPrefix(remainingValue, possibleBase)
	case possibleBase == "today":
		remainingValue = strings.TrimPrefix(remainingValue, possibleBase)
		fallthrough
	default:
		base = t.today()
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

func (t TagExpansion) guessBase(v string) string {
	splitIdx := strings.IndexAny(v, "+-")
	if splitIdx == -1 {
		return v
	}
	return v[:splitIdx]
}

func (t TagExpansion) isWeekday(v string) bool {
	_, ok := weekdayLookup[v]
	return ok
}

func (t TagExpansion) findNext(day time.Weekday) time.Time {
	today := t.today()
	switch {
	case day >= today.Weekday():
		return today.Add(time.Duration(day-today.Weekday()) * 24 * time.Hour)
	default:
		return today.Add(time.Duration(7-today.Weekday()+day) * 24 * time.Hour)
	}
}

func (t TagExpansion) expandInt(list *todotxt.List, i *todotxt.Item, tag string, value string) string {
	var base int
	var remainingValue string
	switch {
	case strings.HasPrefix(value, "max"):
		base = t.globalMaxValue(list, tag)
		remainingValue = strings.TrimPrefix(value, "max")
	case strings.HasPrefix(value, "min"):
		base = t.globalMinValue(list, tag)
		remainingValue = strings.TrimPrefix(value, "min")
	case strings.HasPrefix(value, "pmax"):
		base = t.projectMaxValue(list, i.Projects(), tag)
		remainingValue = strings.TrimPrefix(value, "pmax")
	case strings.HasPrefix(value, "pmin"):
		base = t.projectMinValue(list, i.Projects(), tag)
		remainingValue = strings.TrimPrefix(value, "pmin")
	default:
		// Assume that this is already a properly formatted integer.
		// If it is not this error will be caught by validation.
		return value
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

func (t TagExpansion) globalMaxValue(list *todotxt.List, tag string) int {
	var maximum int = math.MinInt
	var foundOne bool
	for _, i := range list.Tasks() {
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

func (t TagExpansion) globalMinValue(list *todotxt.List, tag string) int {
	var minimum int = math.MaxInt
	var foundOne bool
	for _, i := range list.Tasks() {
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

func (t TagExpansion) projectMaxValue(list *todotxt.List, projects []todotxt.Project, tag string) int {
	var maximum int = math.MinInt
	var foundOne bool
	for _, i := range list.Tasks() {
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

func (t TagExpansion) projectMinValue(list *todotxt.List, projects []todotxt.Project, tag string) int {
	var minimum int = math.MaxInt
	var foundOne bool
	for _, i := range list.Tasks() {
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

func (t TagExpansion) OnValidate(list *todotxt.List, event todotxt.ValidationEvent) error {
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

func (t TagExpansion) today() time.Time {
	now := t.now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
