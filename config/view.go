package config

import (
	"fmt"

	"github.com/Fabian-G/quest/view"
	"github.com/spf13/viper"
)

type OutputMode = string

const (
	JsonOutput        OutputMode = "json"
	InteractiveOutput OutputMode = "interactive"
	ListOutput        OutputMode = "list"
)

var fallbackListViewDef = ViewDef{
	Name:              "list",
	DefaultQuery:      "",
	DefaultProjection: view.StarProjection,
	DefaultSortOrder:  "+done,+creation,+description",
	DefaultOutputMode: InteractiveOutput,
	DefaultClean:      nil,
}

type ViewDef struct {
	Name              string
	DefaultQuery      string
	DefaultProjection string
	DefaultSortOrder  string
	DefaultOutputMode OutputMode
	DefaultClean      []string
}

func DefaultViewDef() ViewDef {
	defView := viper.Sub("default-view")
	if defView == nil {
		return fallbackListViewDef
	}
	return getViewDef(defView)
}

func GetViewDefs() []ViewDef {
	views := viper.Get("view").([]any)
	defs := make([]ViewDef, 0, len(views))
	for idx := range views {
		defs = append(defs, getViewDef(viper.Sub(fmt.Sprintf("view.%d", idx))))
	}
	return defs
}

func getViewDef(subCfg *viper.Viper) ViewDef {
	subCfg.SetDefault("name", "")
	subCfg.SetDefault("query", "")
	subCfg.SetDefault("projection", view.StarProjection)
	subCfg.SetDefault("sortOrder", "+done,+creation,+description")
	subCfg.SetDefault("output", InteractiveOutput)
	subCfg.SetDefault("clean", nil)
	return ViewDef{
		Name:              subCfg.GetString("name"),
		DefaultQuery:      subCfg.GetString("query"),
		DefaultProjection: subCfg.GetString("projection"),
		DefaultSortOrder:  subCfg.GetString("sortOrder"),
		DefaultOutputMode: subCfg.GetString("output"),
		DefaultClean:      subCfg.GetStringSlice("clean"),
	}
}
