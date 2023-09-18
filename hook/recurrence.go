package hook

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Fabian-G/quest/todotxt"
)

var ErrNoRecurrenceBase = errors.New("when the recurrence tag is set, either the due tag or the threshold tag (or both) must be set")

type recurrenceParams struct {
	base      *todotxt.Item
	due       time.Time
	threshold time.Time
	duration  time.Duration
	relative  bool
}

type Recurrence struct {
	list    *todotxt.List
	tags    RecurrenceTags
	nowFunc func() time.Time
}

type RecurrenceTags struct {
	Rec       string
	Due       string
	Threshold string
}

func NewRecurrence(list *todotxt.List, tags RecurrenceTags) todotxt.Hook {
	return &Recurrence{
		list: list,
		tags: tags,
	}
}

func NewRecurrenceWithNowFunc(list *todotxt.List, tags RecurrenceTags, now func() time.Time) todotxt.Hook {
	return &Recurrence{
		list:    list,
		nowFunc: now,
		tags:    tags,
	}
}

func (r Recurrence) OnMod(event todotxt.ModEvent) error {
	if !event.IsCompleteEvent() || len(event.Current.Tags()[r.tags.Rec]) == 0 {
		return nil
	}
	param, err := r.parseRecurrenceParams(event.Current)
	if err != nil {
		return err
	}

	if param.relative {
		return r.spawnRelative(param)
	}
	return r.spawnAbsolute(param)
}

func (r Recurrence) OnValidate(event todotxt.ValidationEvent) error {
	if len(event.Item.Tags()[r.tags.Rec]) == 0 {
		return nil
	}
	_, err := r.parseRecurrenceParams(event.Item)
	if err != nil {
		return fmt.Errorf("recurrent task contains error: %w", err)
	}
	return nil
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
		err := newItem.SetTag(r.tags.Threshold, newThreshold.Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new threshold date when trying to spawn new recurrent task")
		}
		diff := max(params.due.Sub(params.threshold), 0)
		err = newItem.SetTag(r.tags.Due, newThreshold.Add(diff).Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new due date when trying to spawn new recurrent task")
		}
	case params.threshold != zeroTime:
		newThreshold := completionDate.Add(params.duration)
		err := newItem.SetTag(r.tags.Threshold, newThreshold.Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new threshold date when trying to spawn new recurrent task")
		}
	case params.due != zeroTime:
		newDue := completionDate.Add(params.duration)
		err := newItem.SetTag(r.tags.Due, newDue.Format(time.DateOnly))
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
		err := newItem.SetTag(r.tags.Due, params.due.Add(params.duration).Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new due date when trying to spawn new recurrent task")
		}
	}
	if params.threshold != zeroTime {
		err := newItem.SetTag(r.tags.Threshold, params.threshold.Add(params.duration).Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new threshold date when trying to spawn new recurrent task")
		}
	}
	return r.list.Add(newItem)
}

func (r Recurrence) parseRecurrenceParams(current *todotxt.Item) (recurrenceParams, error) {
	params := recurrenceParams{
		base: current,
	}
	tags := current.Tags()
	recTag := r.tags.Rec
	rec := tags[recTag][0] // Cannot be 0 length, because it was checked before
	due := tags[r.tags.Due]
	t := tags[r.tags.Threshold]
	if len(due) == 0 && len(t) == 0 {
		return params, ErrNoRecurrenceBase
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

	if strings.HasPrefix(rec, "+") {
		params.relative = true
		rec = rec[1:]
	}
	duration, err := parseDuration(rec)
	if err != nil {
		return params, fmt.Errorf("could not parse duration %s: %w", rec, err)
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
	case "daily", "d":
		return 24 * time.Hour, nil
	case "weekly", "w":
		return 7 * 24 * time.Hour, nil
	case "monthly", "m":
		return 30 * 24 * time.Hour, nil
	case "yearly", "y":
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
