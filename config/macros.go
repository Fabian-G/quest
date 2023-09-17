package config

import (
	"fmt"
	"log"

	"github.com/Fabian-G/quest/query"
	"github.com/spf13/viper"
)

func registerMacros() {
	macros := viper.Get("macro").([]any)
	for m := range macros {
		setMacroDefaults(m)

		macroDef := fmt.Sprintf("macro.%d", m)
		name := viper.GetString(macroDef + ".name")
		queryS := viper.GetString(macroDef + ".query")
		inTypes := toDTypeSlice(viper.GetStringSlice(macroDef + ".args"))
		outType := query.DType(viper.GetString(macroDef + ".result"))
		injectIt := viper.GetBool(macroDef + ".injectIt")
		err := query.RegisterMacro(name, queryS, inTypes, outType, injectIt)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func setMacroDefaults(idx int) {
	macroDef := fmt.Sprintf("macro.%d", idx)
	viper.SetDefault(macroDef+".result", query.QBool)
	viper.SetDefault(macroDef+"injectIt", false)
}

func toDTypeSlice(s []string) []query.DType {
	d := make([]query.DType, 0, len(s))
	for _, t := range s {
		d = append(d, query.DType(t))
	}
	return d
}
