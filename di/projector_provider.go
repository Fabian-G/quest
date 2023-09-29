package di

import (
	"github.com/Fabian-G/quest/qprojection"
	"github.com/Fabian-G/quest/qscore"
)

func buildProjector(c Config, view ViewDef, calc qscore.Calculator) qprojection.Projector {
	return qprojection.Projector{
		Clean:         view.Clean,
		ScoreCalc:     calc,
		HumanizedTags: c.HumanizedTags(),
		TagTypes:      c.TagTypes(),
	}
}
