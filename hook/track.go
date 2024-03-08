package hook

import (
	"errors"
	"fmt"
	"slices"

	"github.com/Fabian-G/quest/todotxt"
)

var ErrNoActiveTracking = errors.New("No active tracking")

type Tracker interface {
	Start(tags []string) error
	Stop() error
	ActiveTags() ([]string, error)
	SetTags(tags []string) error
}

func NewTracking(tag string, tracker Tracker) *Tracking {
	return &Tracking{
		Tracker: tracker,
		Tag:     tag,
	}
}

type Tracking struct {
	Tracker           Tracker
	Tag               string
	TrimProjectPrefix bool
	TrimContextPrefix bool
	IncludeTags       []string
}

func (t Tracking) OnMod(list *todotxt.List, event todotxt.ModEvent) error {
	prevTracked := event.Previous != nil && !event.Previous.Done() && t.isTrackedTask(event.Previous)
	curTracked := event.Current != nil && !event.Current.Done() && t.isTrackedTask(event.Current)
	switch {
	case !prevTracked && !curTracked:
		// Nothing interesting in this case
		return nil
	case !prevTracked && curTracked:
		t.clearTrackingTag(list, event.Current)
		return t.Tracker.Start(t.trackingTags(event.Current))
	case prevTracked && !curTracked:
		active, err := t.stillActive(event.Previous)
		if err != nil {
			return err
		}
		if active {
			return t.Tracker.Stop()
		}
	case prevTracked && curTracked:
		active, err := t.stillActive(event.Previous)
		if err != nil {
			return err
		}
		if active {
			return t.Tracker.SetTags(t.trackingTags(event.Current))
		} else if event.Previous.Tags()[t.Tag][0] != event.Current.Tags()[t.Tag][0] {
			return t.Tracker.Start(t.trackingTags(event.Current))
		}
	}
	return nil
}

func (t Tracking) stillActive(item *todotxt.Item) (bool, error) {
	tTags := t.trackingTags(item)
	activeTags, err := t.Tracker.ActiveTags()
	if errors.Is(err, ErrNoActiveTracking) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("Could not obtain active tracking tags: %w", err)
	}
	if len(tTags) != len(activeTags) {
		return false, nil
	}
	slices.Sort(tTags)
	slices.Sort(activeTags)
	for i := range tTags {
		if tTags[i] != activeTags[i] {
			return false, nil
		}
	}
	return true, nil
}

func (t Tracking) isTrackedTask(item *todotxt.Item) bool {
	_, ok := item.Tags()[t.Tag]
	return ok
}

func (t Tracking) clearTrackingTag(list *todotxt.List, item *todotxt.Item) {
	for _, task := range list.Tasks() {
		if task == item {
			continue
		}
		// This can not error, because hooks are disabled at this point
		if err := task.SetTag(t.Tag, ""); err != nil {
			panic(err)
		}
	}
}

func (t Tracking) trackingTags(item *todotxt.Item) []string {
	projects := item.Projects()
	contexts := item.Contexts()
	tags := item.Tags()
	description := item.CleanDescription(projects, contexts, tags.Keys())
	trackingTags := make([]string, 0, len(projects)+len(contexts)+1)
	for _, tTag := range t.IncludeTags {
		values, ok := tags[tTag]
		if !ok {
			continue
		}
		for _, v := range values {
			trackingTags = append(trackingTags, fmt.Sprintf("%s:%s", tTag, v))
		}
	}
	for _, p := range projects {
		if t.TrimProjectPrefix {
			trackingTags = append(trackingTags, p.String()[1:])
		} else {
			trackingTags = append(trackingTags, p.String())
		}
	}
	for _, c := range contexts {
		if t.TrimContextPrefix {
			trackingTags = append(trackingTags, c.String()[1:])
		} else {
			trackingTags = append(trackingTags, c.String())
		}
	}
	trackingTags = append(trackingTags, description)
	return trackingTags
}

func (Tracking) OnValidate(list *todotxt.List, event todotxt.ValidationEvent) error {
	return nil
}
