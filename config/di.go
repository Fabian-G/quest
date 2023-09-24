package config

import (
	"log"

	"github.com/Fabian-G/quest/qscore"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/viper"
)

type Di struct {
	ConfigFile           string
	config               *viper.Viper
	repo                 *todotxt.Repo
	doneRepo             *todotxt.Repo
	tagTypes             map[string]qselect.DType
	defaultView          *ViewDef
	viewDefs             []ViewDef
	macros               []MacroDef
	questScoreCalculator *qscore.Calculator
}

func (d *Di) TodoTxtRepo() *todotxt.Repo {
	if d.repo == nil {
		d.repo = buildTodoTxtRepo(d.Config(), d.TagTypes())
	}
	return d.repo
}

func (d *Di) DoneTxtRepo() *todotxt.Repo {
	if d.doneRepo == nil {
		d.doneRepo = buildDoneTxtRepo(d.Config(), d.TagTypes())
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

func (d *Di) DefaultViewDef() ViewDef {
	if d.defaultView == nil {
		defViewDef := buildDefaultViewDef(d.Config())
		d.defaultView = &defViewDef
	}
	return *d.defaultView
}

func (d *Di) ViewDefs() []ViewDef {
	if d.viewDefs == nil {
		d.viewDefs = buildViewDefs(d.Config())
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
