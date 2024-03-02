package hook

import (
	"errors"
	"fmt"
	"slices"

	"github.com/Fabian-G/quest/todotxt"
)

var TrackingTag = "quest-tr"

var ErrNoActiveTracking = errors.New("No active tracking")

type Tracker interface {
	Start(tags []string) error
	Stop() error
	ActiveTags() ([]string, error)
	SetTags(tags []string) error
}

func NewTracking(tracker Tracker) *Tracking {
	return &Tracking{
		tracker: tracker,
	}
}

type Tracking struct {
	tracker           Tracker
	TrimProjectPrefix bool
	TrimContextPrefix bool
}

func (t Tracking) OnMod(list *todotxt.List, event todotxt.ModEvent) error {
	prevTracked := event.Previous != nil && !event.Previous.Done() && isTrackedTask(event.Previous)
	curTracked := event.Current != nil && !event.Current.Done() && isTrackedTask(event.Current)
	switch {
	case !prevTracked && !curTracked:
		// Nothing interesting in this case
		return nil
	case !prevTracked && curTracked:
		clearTrackingTag(list, event.Current)
		return t.tracker.Start(t.trackingTags(event.Current))
	case prevTracked && !curTracked:
		active, err := t.stillActive(event.Previous)
		if err != nil {
			return err
		}
		if active {
			return t.tracker.Stop()
		}
	case prevTracked && curTracked:
		active, err := t.stillActive(event.Previous)
		if err != nil {
			return err
		}
		if active {
			return t.tracker.SetTags(t.trackingTags(event.Current))
		} else if event.Previous.Tags()[TrackingTag][0] != event.Current.Tags()[TrackingTag][0] {
			return t.tracker.Start(t.trackingTags(event.Current))
		}
	}
	return nil
}

func (t Tracking) stillActive(item *todotxt.Item) (bool, error) {
	tTags := t.trackingTags(item)
	activeTags, err := t.tracker.ActiveTags()
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

func isTrackedTask(item *todotxt.Item) bool {
	_, ok := item.Tags()[TrackingTag]
	return ok
}

func clearTrackingTag(list *todotxt.List, item *todotxt.Item) {
	for _, t := range list.Tasks() {
		if t == item {
			continue
		}
		// This can not error, because hooks are disabled at this point
		if err := t.SetTag(TrackingTag, ""); err != nil {
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
