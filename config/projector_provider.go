package config

import (
	"github.com/Fabian-G/quest/qprojection"
	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/qselect"
)

func buildProjector(view ViewDef, humanizedTags []string, tagTypes map[string]qselect.DType, calc qscore.Calculator) qprojection.Projector {
	return qprojection.Projector{
		Clean:         view.DefaultClean,
		ScoreCalc:     calc,
		HumanizedTags: humanizedTags,
		TagTypes:      tagTypes,
	}
}
