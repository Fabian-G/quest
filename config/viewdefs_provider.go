package config

import (
	"fmt"
	"strings"

	"github.com/Fabian-G/quest/qprojection"
	"github.com/spf13/viper"
)

var fallbackListViewDef = ViewDef{
	Name:              "list",
	DefaultQuery:      "",
	DefaultProjection: qprojection.StarProjection,
	DefaultSortOrder:  "+done,+creation,+description",
	DefaultClean:      nil,
}

type ViewDef struct {
	Name              string
	DefaultQuery      string
	DefaultProjection string
	DefaultSortOrder  string
	DefaultClean      []string
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
	return getViewDef(defView)
}

func buildViewDefs(v *viper.Viper) []ViewDef {
	views := v.Get("view").([]any)
	defs := make([]ViewDef, 0, len(views))
	for idx := range views {
		defs = append(defs, getViewDef(v.Sub(fmt.Sprintf("view.%d", idx))))
	}
	return defs
}

func getViewDef(subCfg *viper.Viper) ViewDef {
	subCfg.SetDefault("name", "")
	subCfg.SetDefault("query", "")
	subCfg.SetDefault("projection", qprojection.StarProjection)
	subCfg.SetDefault("sortOrder", "+done,-creation,+description")
	subCfg.SetDefault("clean", nil)
	return ViewDef{
		Name:              subCfg.GetString("name"),
		DefaultQuery:      subCfg.GetString("query"),
		DefaultProjection: subCfg.GetString("projection"),
		DefaultSortOrder:  subCfg.GetString("sortOrder"),
		DefaultClean:      strings.Split(subCfg.GetString("clean"), ","),
		Add:               getAddDef(subCfg.Sub("add")),
	}
}

func getAddDef(subCfg *viper.Viper) AddDef {
	if subCfg == nil {
		return AddDef{}
	}
	subCfg.SetDefault("prefix", "")
	subCfg.SetDefault("suffix", "")
	return AddDef{
		Prefix: subCfg.GetString("prefix"),
		Suffix: subCfg.GetString("suffix"),
	}
}
