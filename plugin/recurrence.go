package plugin

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Fabian-G/todotxt/todotxt"
)

var defaultRecurrenceTag = "rec"
var defaultDueTag = "due"
var defaultThresholdTag = "t"

var ErrNoRecurrenceBase = errors.New("when the recurrence tag is set, either the due tag or the threshold tag (or both) must be set")
var errRecurrenceNotRelevant = errors.New("event nut relevant for recurrence")

type recurrenceParams struct {
	base      *todotxt.Item
	due       time.Time
	threshold time.Time
	duration  time.Duration
	relative  bool
}

type Recurrence struct {
	list    *todotxt.List
	nowFunc func() time.Time
}

func NewRecurrence(list *todotxt.List) *Recurrence {
	return &Recurrence{
		list: list,
	}
}

func NewRecurrenceWithNowFunc(list *todotxt.List, now func() time.Time) *Recurrence {
	return &Recurrence{
		list:    list,
		nowFunc: now,
	}
}

func (r Recurrence) Handle(event todotxt.ModEvent) error {
	param, err := r.parseEvent(event)
	if errors.Is(err, errRecurrenceNotRelevant) {
		return nil
	}
	if err != nil {
		return err
	}

	if param.relative {
		return r.spawnRelative(param)
	}
	return r.spawnAbsolute(param)
}

func (r Recurrence) spawnRelative(params recurrenceParams) error {
	newItem, err := todotxt.BuildItem(
		todotxt.CopyOf(params.base),
		todotxt.WithDone(false),
		todotxt.WithCreationDate(r.now()),
		todotxt.WithoutCompletionDate(),
	)
	if err != nil {
		return fmt.Errorf("could not spawn new recurrent task: %w", err)
	}
	var zeroTime = time.Time{}
	var completionDate time.Time
	if params.base.CompletionDate() != nil {
		completionDate = *params.base.CompletionDate()
	} else {
		completionDate = r.now()
	}

	switch {
	case params.threshold != zeroTime && params.due != zeroTime:
		newThreshold := completionDate.Add(params.duration)
		err := newItem.SetTag(r.thresholdTag(), newThreshold.Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new threshold date when trying to spawn new recurrent task")
		}
		diff := max(params.due.Sub(params.threshold), 0)
		err = newItem.SetTag(r.dueTag(), newThreshold.Add(diff).Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new due date when trying to spawn new recurrent task")
		}
	case params.threshold != zeroTime:
		newThreshold := completionDate.Add(params.duration)
		err := newItem.SetTag(r.thresholdTag(), newThreshold.Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new threshold date when trying to spawn new recurrent task")
		}
	case params.due != zeroTime:
		newDue := completionDate.Add(params.duration)
		err := newItem.SetTag(r.dueTag(), newDue.Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new due date when trying to spawn new recurrent task")
		}
	}
	return r.list.Add(newItem)
}

func (r Recurrence) spawnAbsolute(params recurrenceParams) error {
	newItem, err := todotxt.BuildItem(
		todotxt.CopyOf(params.base),
		todotxt.WithDone(false),
		todotxt.WithCreationDate(r.now()),
		todotxt.WithoutCompletionDate(),
	)
	if err != nil {
		return fmt.Errorf("could not spawn new recurrent task: %w", err)
	}
	var zeroTime = time.Time{}
	if params.due != zeroTime {
		err := newItem.SetTag(r.dueTag(), params.due.Add(params.duration).Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new due date when trying to spawn new recurrent task")
		}
	}
	if params.threshold != zeroTime {
		err := newItem.SetTag(r.thresholdTag(), params.threshold.Add(params.duration).Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new threshold date when trying to spawn new recurrent task")
		}
	}
	return r.list.Add(newItem)
}

func (r Recurrence) recTag() string {
	// TODO Read from config
	return defaultRecurrenceTag
}

func (r Recurrence) dueTag() string {
	// TODO Read from config
	return defaultDueTag
}

func (r Recurrence) thresholdTag() string {
	// TODO Read from config
	return defaultThresholdTag
}

func (r Recurrence) parseEvent(event todotxt.ModEvent) (recurrenceParams, error) {
	params := recurrenceParams{
		base: event.Current,
	}
	tags := event.Current.Tags()
	recTag := r.recTag()
	rec := tags[recTag]
	if len(rec) == 0 {
		return params, errRecurrenceNotRelevant
	}
	due := tags[r.dueTag()]
	t := tags[r.thresholdTag()]
	if len(due) == 0 && len(t) == 0 {
		return params, ErrNoRecurrenceBase
	}
	if event.Previous == nil || event.Current == nil {
		return params, errRecurrenceNotRelevant
	}
	if event.Previous.Done() || !event.Current.Done() {
		return params, errRecurrenceNotRelevant
	}

	if len(due) > 0 {
		dueDate, err := time.Parse(time.DateOnly, due[0])
		if err != nil {
			return params, fmt.Errorf("could not parse due date %s: %w", due[0], err)
		}
		params.due = dueDate
	}
	if len(t) > 0 {
		thresholdDate, err := time.Parse(time.DateOnly, t[0])
		if err != nil {
			return params, fmt.Errorf("could not parse threshold date %s: %w", t[0], err)
		}
		params.threshold = thresholdDate
	}

	recTagValue := rec[0]
	if strings.HasPrefix(recTagValue, "+") {
		params.relative = true
		recTagValue = recTagValue[1:]
	}
	duration, err := parseDuration(recTagValue)
	if err != nil {
		return params, fmt.Errorf("could not parse duration %s: %w", recTagValue, err)
	}
	params.duration = duration.Abs().Round(24 * time.Hour)
	return params, nil
}

func (r Recurrence) now() time.Time {
	if r.nowFunc != nil {
		return r.nowFunc()
	}
	return time.Now()
}

var numRegex = regexp.MustCompile("[0-9]+")

var unitMap = map[string]time.Duration{
	"d":      24 * time.Hour,
	"days":   24 * time.Hour,
	"w":      7 * 24 * time.Hour,
	"weeks":  7 * 24 * time.Hour,
	"m":      30 * 24 * time.Hour,
	"months": 30 * 24 * time.Hour,
	"y":      365 * 24 * time.Hour,
	"years":  365 * 24 * time.Hour,
}

func parseDuration(duration string) (time.Duration, error) {
	switch duration {
	case "daily":
		return 24 * time.Hour, nil
	case "weekly":
		return 7 * 24 * time.Hour, nil
	case "monthly":
		return 30 * 24 * time.Hour, nil
	case "yearly":
		return 365 * 24 * time.Hour, nil
	default:
		valueIdx := numRegex.FindStringIndex(duration)
		if valueIdx == nil {
			return 0, errors.New("missing value in duration expression")
		}
		if valueIdx[1] == len(duration) {
			return 0, errors.New("missing unit")
		}
		unit, ok := unitMap[duration[valueIdx[1]:]]
		if !ok {
			return 0, fmt.Errorf("unknown unit %s", duration[valueIdx[1]:])
		}
		value, err := strconv.Atoi(duration[valueIdx[0]:valueIdx[1]])
		if err != nil {
			return 0, err
		}
		return time.Duration(value) * unit, nil
	}
}
