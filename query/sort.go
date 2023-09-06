package query

import (
	"fmt"
	"strings"

	"github.com/Fabian-G/quest/todotxt"
)

type sortOrder int

const (
	asc  = 1
	desc = -1
)

func SortFunc(sort string) (func(*todotxt.Item, *todotxt.Item) int, error) {
	sortingKeys := strings.Split(sort, ",")
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
		case "description":
			compareFuncs = append(compareFuncs, order.compareDescription)
		default:
			if !strings.HasPrefix(key, "tag:") {
				return nil, fmt.Errorf("unknown sort key %s", key)
			}
			parts := strings.Split(key, ":")
			if len(parts) < 2 {
				return nil, fmt.Errorf("when sorting by tag a tag name must be specified e.g. tag:rec")
			}
			tagKey := parts[1]
			compareFuncs = append(compareFuncs, order.compareTag(tagKey))
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

func (o sortOrder) compareCreation(i1 *todotxt.Item, i2 *todotxt.Item) int {
	switch {
	case i1.CreationDate() == nil && i2.CreationDate() != nil:
		return -1
	case i1.CreationDate() != nil && i2.CreationDate() == nil:
		return 1
	case i1.CreationDate() == nil && i2.CreationDate() == nil:
		return 0
	default:
		return int(o) * i1.CreationDate().Compare(*i2.CreationDate())
	}
}

func (o sortOrder) compareCompletion(i1 *todotxt.Item, i2 *todotxt.Item) int {
	switch {
	case i1.CompletionDate() == nil && i2.CompletionDate() != nil:
		return -1
	case i1.CompletionDate() != nil && i2.CompletionDate() == nil:
		return 1
	case i1.CompletionDate() == nil && i2.CompletionDate() == nil:
		return 0
	default:
		return int(o) * i1.CompletionDate().Compare(*i2.CompletionDate())
	}
}

func (o sortOrder) compareTag(tagKey string) func(*todotxt.Item, *todotxt.Item) int {
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
			// TODO Lookup proper comparison function depending on the tag type
			return int(o) * strings.Compare(firstTags[0], secondTags[0])
		}
	}
}

func (o sortOrder) compareDescription(i1 *todotxt.Item, i2 *todotxt.Item) int {
	return int(o) * strings.Compare(i1.Description(), i2.Description())
}
func orderFactor(key string) sortOrder {
	if strings.HasPrefix(key, "-") {
		return desc
	}
	return asc
}
