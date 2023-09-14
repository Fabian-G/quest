package config

import (
	"log"
	"strings"

	"github.com/Fabian-G/quest/view"
	"github.com/spf13/viper"
)

type OutputMode = string

const (
	JsonOutput        OutputMode = "json"
	InteractiveOutput OutputMode = "interactive"
	ListOutputMode    OutputMode = "list"
)

var ListViewDef = ViewDef{
	Name:              "list",
	DefaultSelection:  "",
	DefaultProjection: view.StarProjection,
	DefaultSortOrder:  "+done,+creation,+description",
	DefaultOutputMode: InteractiveOutput,
	DefaultClean:      nil,
}

type ViewDef struct {
	Name              string
	DefaultSelection  string
	DefaultProjection string
	DefaultSortOrder  string
	DefaultOutputMode OutputMode
	DefaultClean      []string
}

func GetViewDefs() []ViewDef {
	views := viper.GetStringMap("view")
	defs := make([]ViewDef, 0, len(views))
	for name, viewDefA := range views {
		viewDefM, ok := viewDefA.(map[string]any)
		if !ok {
			log.Fatalf("error in config file. expected view definition in section [view.%s], but got %T", name, viewDefA)
		}
		var (
			selection  string   = ""
			projection string   = view.StarProjection
			sortOrder  string   = "+done,+creation,+description"
			output     string   = InteractiveOutput
			clean      []string = nil
		)
		if s, ok := viewDefM["query"]; ok {
			selection = s.(string)
		}
		if p, ok := viewDefM["projection"]; ok {
			projection = p.(string)
		}
		if s, ok := viewDefM["sort"]; ok {
			sortOrder = s.(string)
		}
		if i, ok := viewDefM["output"]; ok {
			output = i.(string)
		}
		if c, ok := viewDefM["clean"]; ok {
			cleanS := c.(string)
			clean = strings.Split(cleanS, ",")
		}
		defs = append(defs, ViewDef{
			Name:              name,
			DefaultSelection:  selection,
			DefaultProjection: projection,
			DefaultSortOrder:  sortOrder,
			DefaultOutputMode: output,
			DefaultClean:      clean,
		})
	}
	return defs
}
