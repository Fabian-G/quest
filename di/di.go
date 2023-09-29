package di

import (
	"log"

	"github.com/Fabian-G/quest/qprojection"
	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/qsort"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type Container struct {
	ConfigFile           string
	config               *Config
	repo                 *todotxt.Repo
	doneRepo             *todotxt.Repo
	questScoreCalculator *qscore.Calculator
	sortCompiler         *qsort.Compiler
	projector            map[string]*qprojection.Projector
	editor               Editor
}

func (d *Container) TodoTxtRepo() *todotxt.Repo {
	if d.repo == nil {
		d.repo = buildTodoTxtRepo(d.Config(), d.SortCompiler())
	}
	return d.repo
}

func (d *Container) DoneTxtRepo() *todotxt.Repo {
	if d.doneRepo == nil {
		d.doneRepo = buildDoneTxtRepo(d.Config(), d.SortCompiler())
	}
	return d.doneRepo
}

func (d *Container) Config() Config {
	if d.config == nil {
		var err error
		config, err := buildConfig(d.ConfigFile)
		if err != nil {
			log.Fatal(err)
		}
		d.config = &config
	}
	return *d.config
}

func (d *Container) QuestScoreCalculator() qscore.Calculator {
	if d.questScoreCalculator == nil {
		calc := buildScoreCalculator(d.Config())
		d.questScoreCalculator = &calc
	}
	return *d.questScoreCalculator
}

func (d *Container) SortCompiler() qsort.Compiler {
	if d.sortCompiler == nil {
		sort := buildSortCompiler(d.Config(), d.QuestScoreCalculator())
		d.sortCompiler = &sort
	}
	return *d.sortCompiler
}

func (d *Container) Projector(cmd *cobra.Command) qprojection.Projector {
	view := cmd.Name()
	if d.projector == nil {
		d.projector = make(map[string]*qprojection.Projector)
	}
	if _, ok := d.projector[view]; !ok {
		config := d.Config()
		viewDef, ok := config.Views[view]
		if !ok {
			viewDef = config.DefaultView
		}
		projectorForView := buildProjector(config, viewDef, d.QuestScoreCalculator())
		d.projector[view] = &projectorForView
	}
	return *d.projector[view]
}

func (d *Container) Editor() Editor {
	if d.editor == nil {
		d.editor = buildEditor(d.Config())
	}
	return d.editor
}

func (d *Container) SetConfig(c Config) {
	d.config = &c
}

func (d *Container) SetEditor(e Editor) {
	d.editor = e
}
