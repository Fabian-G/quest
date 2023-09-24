package config

import (
	"fmt"
	"log"

	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/viper"
)

func buildScoreCalculator(v *viper.Viper) qscore.Calculator {
	minPriority, err := todotxt.PriorityFromString(v.GetString("min-priority"))
	if err != nil {
		log.Fatal(fmt.Errorf("could not parse min priority for quest-score: %w", err))
	}
	scoreConfig := v.Sub(QuestScore)
	return qscore.Calculator{
		UrgencyTag:   scoreConfig.GetString("urgency-tag"),
		UrgencyBegin: scoreConfig.GetInt("urgency-begin"),
		MinPriority:  minPriority,
	}
}
