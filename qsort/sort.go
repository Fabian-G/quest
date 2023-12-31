package qsort

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Fabian-G/quest/qduration"
	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
)

type Compiler struct {
	TagTypes        map[string]qselect.DType
	ScoreCalculator qscore.Calculator
}

func (c Compiler) CompileSortFunc(sortingKeys []string) (func(*todotxt.Item, *todotxt.Item) int, error) {
	compareFuncs := make([]func(*todotxt.Item, *todotxt.Item) int, 0, len(sortingKeys))

	for _, key := range sortingKeys {
		key = strings.TrimSpace(key)
		order := orderFactor(key)
		key = strings.TrimLeft(key, "+-")

		switch key {
		case "":
			continue
		case "done":
			compareFuncs = append(compareFuncs, order.compareDone)
		case "creation":
			compareFuncs = append(compareFuncs, order.compareCreation)
		case "completion":
			compareFuncs = append(compareFuncs, order.compareCompletion)
		case "priority":
			compareFuncs = append(compareFuncs, order.comparePriority)
		case "description":
			compareFuncs = append(compareFuncs, order.compareDescription)
		case "project", "projects":
			compareFuncs = append(compareFuncs, order.compareProject)
		case "context", "contexts":
			compareFuncs = append(compareFuncs, order.compareContext)
		case "score":
			compareFuncs = append(compareFuncs, order.compareScores(c.ScoreCalculator))
		default:
			if !strings.HasPrefix(key, "tag:") {
				return nil, fmt.Errorf("unknown sort key %s", key)
			}
			parts := strings.Split(key, ":")
			if len(parts) < 2 {
				return nil, fmt.Errorf("when sorting by tag a tag name must be specified e.g. tag:rec")
			}
			tagKey := parts[1]
			compareFuncs = append(compareFuncs, order.compareTag(tagKey, c.TagTypes))
		}
	}
	return func(first, second *todotxt.Item) int {
		for _, cFunc := range compareFuncs {
			result := cFunc(first, second)
			if result != 0 {
				return result
			}
		}
		return 0
	}, nil
}

type sortOrder int

const (
	asc  = 1
	desc = -1
)

func (o sortOrder) compareDone(i1 *todotxt.Item, i2 *todotxt.Item) int {
	switch {
	case !i1.Done() && i2.Done():
		return int(o) * -1
	case i1.Done() && !i2.Done():
		return int(o) * 1
	default:
		return 0
	}
}

func (o sortOrder) comparePriority(i1 *todotxt.Item, i2 *todotxt.Item) int {
	prio1 := i1.Priority()
	prio2 := i2.Priority()
	return int(o) * cmp.Compare(prio1, prio2)
}

func (o sortOrder) compareCreation(i1 *todotxt.Item, i2 *todotxt.Item) int {
	return int(o) * compareOptionalsFunc(i1.CreationDate(), i2.CreationDate(), func(t1, t2 time.Time) int {
		return t1.Compare(t2)
	})
}

func (o sortOrder) compareCompletion(i1 *todotxt.Item, i2 *todotxt.Item) int {
	return int(o) * compareOptionalsFunc(i1.CompletionDate(), i2.CompletionDate(), func(t1, t2 time.Time) int {
		return t1.Compare(t2)
	})
}

func (o sortOrder) compareTag(tagKey string, tagTypes map[string]qselect.DType) func(*todotxt.Item, *todotxt.Item) int {
	return func(i1, i2 *todotxt.Item) int {
		firstTags := i1.Tags()[tagKey]
		secondTags := i2.Tags()[tagKey]
		switch {
		case len(firstTags) == 0 && len(secondTags) == 0:
			return 0
		case len(firstTags) == 0 && len(secondTags) != 0:
			return -1
		case len(firstTags) != 0 && len(secondTags) == 0:
			return 1
		default:
			switch tagTypes[tagKey] {
			case qselect.QInt:
				return int(o) * compareOptionals(
					valueOrNil(strconv.Atoi(firstTags[0])),
					valueOrNil(strconv.Atoi(secondTags[0])),
				)
			case qselect.QDuration:
				return int(o) * compareOptionalsFunc(
					valueOrNil(qduration.Parse(firstTags[0])),
					valueOrNil(qduration.Parse(secondTags[0])),
					func(d1, d2 qduration.Duration) int {
						return cmp.Compare(d1.Days(), d2.Days())
					},
				)
			case qselect.QDate:
				return int(o) * compareOptionalsFunc(
					valueOrNil(time.Parse(time.DateOnly, firstTags[0])),
					valueOrNil(time.Parse(time.DateOnly, secondTags[0])),
					func(t1, t2 time.Time) int {
						return t1.Compare(t2)
					},
				)
			default:
				return int(o) * strings.Compare(firstTags[0], secondTags[0])
			}
		}
	}
}

func (o sortOrder) compareDescription(i1 *todotxt.Item, i2 *todotxt.Item) int {
	return int(o) * strings.Compare(i1.Description(), i2.Description())
}

func (o sortOrder) compareScores(calc qscore.Calculator) func(*todotxt.Item, *todotxt.Item) int {
	return func(i1, i2 *todotxt.Item) int {
		score1 := calc.ScoreOf(i1)
		score2 := calc.ScoreOf(i2)
		return int(o) * cmp.Compare(score1.Score, score2.Score)
	}
}

func (o sortOrder) compareProject(i1 *todotxt.Item, i2 *todotxt.Item) int {
	i1Projects := i1.Projects()
	i2Projects := i2.Projects()
	slices.SortFunc(i1Projects, func(a, b todotxt.Project) int {
		return int(o) * strings.Compare(string(a), string(b))
	})
	slices.SortFunc(i2Projects, func(a, b todotxt.Project) int {
		return int(o) * strings.Compare(string(a), string(b))
	})
	var p1 *string
	var p2 *string
	if len(i1Projects) > 0 {
		p1String := string(i1Projects[0])
		p1 = &p1String
	}
	if len(i2Projects) > 0 {
		p2String := string(i2Projects[0])
		p2 = &p2String
	}
	return int(o) * compareOptionals(p1, p2)
}

func (o sortOrder) compareContext(i1 *todotxt.Item, i2 *todotxt.Item) int {
	i1Contexts := i1.Contexts()
	i2Contexts := i2.Contexts()
	slices.SortFunc(i1Contexts, func(a, b todotxt.Context) int {
		return int(o) * strings.Compare(string(a), string(b))
	})
	slices.SortFunc(i2Contexts, func(a, b todotxt.Context) int {
		return int(o) * strings.Compare(string(a), string(b))
	})
	var p1 *string
	var p2 *string
	if len(i1Contexts) > 0 {
		p1String := string(i1Contexts[0])
		p1 = &p1String
	}
	if len(i2Contexts) > 0 {
		p2String := string(i2Contexts[0])
		p2 = &p2String
	}
	return int(o) * compareOptionals(p1, p2)
}

func compareOptionals[T cmp.Ordered](a *T, b *T) int {
	return compareOptionalsFunc(a, b, cmp.Compare)
}

func compareOptionalsFunc[T any](a *T, b *T, compare func(T, T) int) int {
	switch {
	case a == nil && b != nil:
		return -1
	case a != nil && b == nil:
		return 1
	case a == nil && b == nil:
		return 0
	default:
		return compare(*a, *b)
	}
}

func valueOrNil[T any](val T, err error) *T {
	if err != nil {
		return nil
	}
	return &val
}

func orderFactor(key string) sortOrder {
	if strings.HasPrefix(key, "-") {
		return desc
	}
	return asc
}
