package di

import (
	"fmt"
	"log"
	"strings"

	"github.com/Fabian-G/quest/qduration"
	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/todotxt"
)

func buildScoreCalculator(c Config) qscore.Calculator {
	minPriority, err := todotxt.PriorityFromString(c.QuestScore.MinPriority)
	if err != nil {
		log.Fatal(fmt.Errorf("could not parse min priority for quest-score: %w", err))
	}
	urgencyDefault, err := qduration.Parse(c.QuestScore.UrgencyDefault)
	if err != nil {
		log.Fatal(fmt.Errorf("could not parse urgency-default for quest-score. Expected a duration, got: %s. Err: %w", c.QuestScore.UrgencyDefault, err))
	}

	urgencyTags := make([]qscore.UrgencyTag, 0, len(c.QuestScore.UrgencyTags))
	for _, tag := range c.QuestScore.UrgencyTags {
		dividerIdx := strings.LastIndex(tag, "+")
		offset := qduration.Duration{}
		if dividerIdx != -1 {
			offset, err = qduration.Parse(tag[dividerIdx+1:])
			if err != nil {
				log.Fatalf("could not parse urgencyTag offset for tag %s: %s", tag, err)
			}
		}
		if dividerIdx == -1 {
			dividerIdx = len(tag)
		}

		urgencyTags = append(urgencyTags, qscore.UrgencyTag{
			Tag:    tag[:dividerIdx],
			Offset: offset,
		})
	}

	return qscore.Calculator{
		UrgencyTags:    urgencyTags,
		UrgencyBegin:   c.QuestScore.UrgencyBegin,
		DefaultUrgency: urgencyDefault,
		MinPriority:    minPriority,
	}
}
