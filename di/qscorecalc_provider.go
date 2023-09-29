package di

import (
	"fmt"
	"log"

	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/todotxt"
)

func buildScoreCalculator(c Config) qscore.Calculator {
	minPriority, err := todotxt.PriorityFromString(c.QuestScore.MinPriority)
	if err != nil {
		log.Fatal(fmt.Errorf("could not parse min priority for quest-score: %w", err))
	}
	return qscore.Calculator{
		UrgencyTag:   c.QuestScore.UrgencyTag,
		UrgencyBegin: c.QuestScore.UrgencyBegin,
		MinPriority:  minPriority,
	}
}
