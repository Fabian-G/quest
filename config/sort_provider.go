package config

import (
	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/qsort"
)

func buildSortCompiler(tagTypes map[string]qselect.DType, calc qscore.Calculator) qsort.Compiler {
	return qsort.Compiler{
		TagTypes:        tagTypes,
		ScoreCalculator: calc,
	}
}
