package config

import (
	"log"
	"slices"

	"github.com/Fabian-G/quest/qprojection"
	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/qsort"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Di struct {
	ConfigFile           string
	config               *viper.Viper
	repo                 *todotxt.Repo
	doneRepo             *todotxt.Repo
	tagTypes             map[string]qselect.DType
	humanizedTags        []string
	defaultView          *ViewDef
	viewDefs             []ViewDef
	macros               []MacroDef
	questScoreCalculator *qscore.Calculator
	sortCompiler         *qsort.Compiler
	projector            map[string]*qprojection.Projector
	editor               Editor
}

func (d *Di) TodoTxtRepo() *todotxt.Repo {
	if d.repo == nil {
		d.repo = buildTodoTxtRepo(d.Config(), d.SortCompiler(), d.TagTypes())
	}
	return d.repo
}

func (d *Di) DoneTxtRepo() *todotxt.Repo {
	if d.doneRepo == nil {
		d.doneRepo = buildDoneTxtRepo(d.Config(), d.SortCompiler())
	}
	return d.doneRepo
}

func (d *Di) Config() *viper.Viper {
	if d.config == nil {
		var err error
		if d.config, err = buildConfig(d.ConfigFile); err != nil {
			log.Fatal(err)
		}
	}
	return d.config
}

func (d *Di) TagTypes() map[string]qselect.DType {
	if d.tagTypes == nil {
		d.tagTypes = buildTagTypes(d.Config())
	}
	return d.tagTypes
}

func (d *Di) HumanizedTags() []string {
	if d.humanizedTags == nil {
		d.humanizedTags = buildHumanizedTags(d.Config())
	}
	return d.humanizedTags
}

func (d *Di) DefaultViewDef() ViewDef {
	if d.defaultView == nil {
		defViewDef := buildDefaultViewDef(d.Config())
		d.defaultView = &defViewDef
	}
	return *d.defaultView
}

func (d *Di) ViewDefs() []ViewDef {
	if d.viewDefs == nil {
		d.viewDefs = buildViewDefs(d.DefaultViewDef(), d.Config())
	}
	return d.viewDefs
}

func (d *Di) MacroDefs() []MacroDef {
	if d.macros == nil {
		d.macros = buildMacroDefs(d.Config())
	}
	return d.macros
}

func (d *Di) QuestScoreCalculator() qscore.Calculator {
	if d.questScoreCalculator == nil {
		calc := buildScoreCalculator(d.Config())
		d.questScoreCalculator = &calc
	}
	return *d.questScoreCalculator
}

func (d *Di) SortCompiler() qsort.Compiler {
	if d.sortCompiler == nil {
		sort := buildSortCompiler(d.TagTypes(), d.QuestScoreCalculator())
		d.sortCompiler = &sort
	}
	return *d.sortCompiler
}

func (d *Di) Projector(cmd *cobra.Command) qprojection.Projector {
	view := cmd.Name()
	if d.projector == nil {
		d.projector = make(map[string]*qprojection.Projector)
	}
	if _, ok := d.projector[view]; !ok {
		viewDef := d.DefaultViewDef()
		viewDefIdx := slices.IndexFunc[[]ViewDef, ViewDef](d.ViewDefs(), func(vd ViewDef) bool { return vd.Name == view })
		if viewDefIdx != -1 {
			viewDef = d.ViewDefs()[viewDefIdx]
		}
		projectorForView := buildProjector(viewDef, d.HumanizedTags(), d.TagTypes(), d.QuestScoreCalculator())
		d.projector[view] = &projectorForView
	}
	return *d.projector[view]
}

func (d *Di) Editor() Editor {
	if d.editor == nil {
		d.editor = buildEditor(d.Config())
	}
	return d.editor
}

func (d *Di) SetConfig(v *viper.Viper) {
	d.config = v
}

func (d *Di) SetEditor(e Editor) {
	d.editor = e
}
