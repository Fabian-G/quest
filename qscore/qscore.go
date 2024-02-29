package qscore

import (
	"errors"
	"math"
	"time"

	"github.com/Fabian-G/quest/qduration"
	"github.com/Fabian-G/quest/todotxt"
)

const (
	urgencyThreshold    = 3 / 5.0 * 10
	importanceThreshold = 3/5.0*10 - 1
)

type UrgencyTag struct {
	Tag    string
	Offset qduration.Duration
}

type Score struct {
	Score      float32
	Urgency    float32
	Importance float32
}

func (s Score) IsUrgent() bool {
	return s.Urgency >= urgencyThreshold
}

func (s Score) IsImportant() bool {
	return s.Importance >= importanceThreshold
}

type Calculator struct {
	UrgencyTags    []UrgencyTag
	UrgencyBegin   int
	DefaultUrgency qduration.Duration
	MinPriority    todotxt.Priority
	NowFunc        func() time.Time
}

func (c Calculator) ScoreOf(item *todotxt.Item) Score {
	score := Score{}
	if item.Done() {
		return score
	}
	score.Urgency = c.urgency(item)
	score.Importance = c.importance(item)
	score.Score = c.score(score.Urgency, score.Importance)

	return score
}

func (c Calculator) importance(item *todotxt.Item) float32 {
	priority := item.Priority()
	if priority == todotxt.PrioNone || c.MinPriority == todotxt.PrioNone {
		return 0
	}
	if priority <= c.MinPriority {
		return 1
	}

	maxImportance := float32(10.0)
	minImportance := float32(1.0)
	return minImportance + (float32(priority)-float32(c.MinPriority))*(maxImportance-minImportance)/(float32(todotxt.PrioA)-float32(c.MinPriority))
}

func (c Calculator) urgency(item *todotxt.Item) float32 {
	date, err := c.urgencyDate(item)
	if err != nil {
		return 0
	}
	days := float32(date.Sub(c.now()).Hours()) / 24
	if days <= 0 {
		return 10
	}
	if math.Round(float64(days)) >= math.Round(float64(c.UrgencyBegin)) {
		return 1
	}

	maxUrgency := float32(10.0)
	minUrgency := float32(1.0)
	return maxUrgency + days*(minUrgency-maxUrgency)/float32(c.UrgencyBegin)
}

func (c Calculator) urgencyDate(item *todotxt.Item) (time.Time, error) {
	itemTags := item.Tags()
	for _, tag := range c.UrgencyTags {
		if len(itemTags[tag.Tag]) > 0 {
			date, err := time.Parse(time.DateOnly, itemTags[tag.Tag][0])
			if err != nil {
				return time.Time{}, err
			}
			return tag.Offset.AddTo(date), nil
		}
	}
	if item.CreationDate() != nil && c.DefaultUrgency.Days() > 0 {
		return c.DefaultUrgency.AddTo(*item.CreationDate()), nil
	}
	return time.Time{}, errors.New("No urgency date found")
}

func (c Calculator) score(urgency, importance float32) float32 {
	squaredSum := urgency*urgency + importance*importance
	return float32(math.Sqrt(float64(squaredSum / 2)))
}

func (c Calculator) now() time.Time {
	if c.NowFunc == nil {
		return time.Now()
	}
	return c.NowFunc()
}
