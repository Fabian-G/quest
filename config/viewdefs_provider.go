package config

import (
	"fmt"
	"log"

	"github.com/Fabian-G/quest/qprojection"
	"github.com/spf13/viper"
)

var fallbackListViewDef = ViewDef{
	Name:              "list",
	DefaultQuery:      "",
	DefaultProjection: qprojection.StarProjection,
	DefaultSortOrder:  nil,
	DefaultClean:      nil,
}

type ViewDef struct {
	Name              string
	DefaultQuery      string
	DefaultProjection []string
	DefaultSortOrder  []string
	DefaultClean      []string
	Interactive       bool
	Add               AddDef
}

type AddDef struct {
	Prefix string
	Suffix string
}

func buildDefaultViewDef(v *viper.Viper) ViewDef {
	defView := v.Sub("default-view")
	if defView == nil {
		return fallbackListViewDef
	}
	return getViewDef(fallbackListViewDef, defView)
}

func buildViewDefs(defaultView ViewDef, v *viper.Viper) []ViewDef {
	views, ok := v.Get(ViewsKey).([]any)
	if !ok {
		log.Fatal("error in config: expected view to be a list")
	}
	defs := make([]ViewDef, 0, len(views))
	for idx := range views {
		defs = append(defs, getViewDef(defaultView, v.Sub(fmt.Sprintf("view.%d", idx))))
	}
	return defs
}

func getViewDef(defaultView ViewDef, subCfg *viper.Viper) ViewDef {
	subCfg.SetDefault("query", defaultView.DefaultQuery)
	subCfg.SetDefault("projection", defaultView.DefaultProjection)
	subCfg.SetDefault("sort", defaultView.DefaultSortOrder)
	subCfg.SetDefault("clean", defaultView.DefaultClean)
	subCfg.SetDefault("interactive", defaultView.Interactive)
	return ViewDef{
		Name:              subCfg.GetString("name"),
		DefaultQuery:      subCfg.GetString("query"),
		DefaultProjection: subCfg.GetStringSlice("projection"),
		DefaultSortOrder:  subCfg.GetStringSlice("sort"),
		DefaultClean:      subCfg.GetStringSlice("clean"),
		Add:               getAddDef(defaultView, subCfg.Sub("add")),
		Interactive:       subCfg.GetBool("interactive"),
	}
}

func getAddDef(defaultView ViewDef, subCfg *viper.Viper) AddDef {
	if subCfg == nil {
		return defaultView.Add
	}
	subCfg.SetDefault("prefix", defaultView.Add.Prefix)
	subCfg.SetDefault("suffix", defaultView.Add.Suffix)
	return AddDef{
		Prefix: subCfg.GetString("prefix"),
		Suffix: subCfg.GetString("suffix"),
	}
}
