package hook

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Fabian-G/quest/qduration"
	"github.com/Fabian-G/quest/todotxt"
)

var ErrNoRecurrenceBase = errors.New("when the recurrence tag is set, either the due tag or the threshold tag (or both) must be set")

type recurrenceParams struct {
	list      *todotxt.List
	base      *todotxt.Item
	due       time.Time
	threshold time.Time
	duration  qduration.Duration
	relative  bool
}

type Recurrence struct {
	tags    RecurrenceTags
	nowFunc func() time.Time
}

type RecurrenceTags struct {
	Rec       string
	Due       string
	Threshold string
}

func NewRecurrence(tags RecurrenceTags) todotxt.Hook {
	return &Recurrence{
		tags: tags,
	}
}

func NewRecurrenceWithNowFunc(tags RecurrenceTags, now func() time.Time) todotxt.Hook {
	return &Recurrence{
		nowFunc: now,
		tags:    tags,
	}
}

func (r Recurrence) OnMod(list *todotxt.List, event todotxt.ModEvent) error {
	if event.Current == nil {
		return nil
	}
	if !event.IsCompleteEvent() || len(event.Current.Tags()[r.tags.Rec]) == 0 {
		return r.OnValidate(list, todotxt.ValidationEvent{
			Item: event.Current,
		})
	}
	param, err := r.parseRecurrenceParams(list, event.Current)
	if err != nil {
		return err
	}

	if param.relative {
		return r.spawnRelative(param)
	}
	return r.spawnAbsolute(param)
}

func (r Recurrence) OnValidate(list *todotxt.List, event todotxt.ValidationEvent) error {
	if len(event.Item.Tags()[r.tags.Rec]) == 0 {
		return nil
	}
	_, err := r.parseRecurrenceParams(list, event.Item)
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
		newThreshold := params.duration.AddTo(completionDate)
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
		newThreshold := params.duration.AddTo(completionDate)
		err := newItem.SetTag(r.tags.Threshold, newThreshold.Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new threshold date when trying to spawn new recurrent task")
		}
	case params.due != zeroTime:
		newDue := params.duration.AddTo(completionDate)
		err := newItem.SetTag(r.tags.Due, newDue.Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new due date when trying to spawn new recurrent task")
		}
	}
	return params.list.Add(newItem)
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
		err := newItem.SetTag(r.tags.Due, params.duration.AddTo(params.due).Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new due date when trying to spawn new recurrent task")
		}
	}
	if params.threshold != zeroTime {
		err := newItem.SetTag(r.tags.Threshold, params.duration.AddTo(params.threshold).Format(time.DateOnly))
		if err != nil {
			return fmt.Errorf("failed to set new threshold date when trying to spawn new recurrent task")
		}
	}
	return params.list.Add(newItem)
}

func (r Recurrence) parseRecurrenceParams(list *todotxt.List, current *todotxt.Item) (recurrenceParams, error) {
	params := recurrenceParams{
		base: current,
		list: list,
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

	if !strings.HasPrefix(rec, "+") {
		params.relative = true
	}
	duration, err := qduration.Parse(rec)
	if err != nil {
		return params, fmt.Errorf("could not parse duration %s: %w", rec, err)
	}
	params.duration = duration.Abs()
	return params, nil
}

func (r Recurrence) now() time.Time {
	if r.nowFunc != nil {
		return r.nowFunc()
	}
	return time.Now()
}
