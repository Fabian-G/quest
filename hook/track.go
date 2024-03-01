package hook

import (
	"errors"
	"fmt"

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
	tracker Tracker
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
		return t.tracker.Start(trackingTags(event.Current))
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
			return t.tracker.SetTags(trackingTags(event.Current))
		} else if event.Previous.Tags()[TrackingTag][0] != event.Current.Tags()[TrackingTag][0] {
			return t.tracker.Start(trackingTags(event.Current))
		}
	}
	return nil
}

func (t Tracking) stillActive(item *todotxt.Item) (bool, error) {
	tTags := trackingTags(item)
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

func trackingTags(item *todotxt.Item) []string {
	projects := item.Projects()
	contexts := item.Contexts()
	tags := item.Tags()
	description := item.CleanDescription(projects, contexts, tags.Keys())
	trackingTags := make([]string, 0, len(projects)+len(contexts)+1)
	for _, p := range projects {
		trackingTags = append(trackingTags, p.String())
	}
	for _, c := range contexts {
		trackingTags = append(trackingTags, c.String())
	}
	trackingTags = append(trackingTags, description)
	return trackingTags
}

func (Tracking) OnValidate(list *todotxt.List, event todotxt.ValidationEvent) error {
	var trTags int
	for _, item := range list.Tasks() {
		if _, ok := item.Tags()[TrackingTag]; ok {
			trTags++
		}
		if trTags > 1 {
			return fmt.Errorf("Found multiple tracking tags %s.\nThis should never happen unless you edit them manually.", TrackingTag)
		}
	}
	return nil
}
