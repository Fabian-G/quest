package qscore

import (
	"math"
	"time"

	"github.com/Fabian-G/quest/todotxt"
)

const (
	urgencyThreshold    = 3 / 5.0 * 10
	importanceThreshold = 3/5.0*10 - 1
)

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
	UrgencyTag   string
	UrgencyBegin int
	MinPriority  todotxt.Priority
	NowFunc      func() time.Time
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
	dateTag := item.Tags()[c.UrgencyTag]
	if len(dateTag) == 0 {
		return 0
	}
	date, err := time.Parse(time.DateOnly, dateTag[0])
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
