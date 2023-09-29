package di

import (
	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/qsort"
)

func buildSortCompiler(c Config, calc qscore.Calculator) qsort.Compiler {
	return qsort.Compiler{
		TagTypes:        c.TagTypes(),
		ScoreCalculator: calc,
	}
}
