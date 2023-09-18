package config

import (
	"fmt"
	"log"

	"github.com/Fabian-G/quest/qselect"
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
		outType := qselect.DType(viper.GetString(macroDef + ".result"))
		injectIt := viper.GetBool(macroDef + ".injectIt")
		err := qselect.RegisterMacro(name, queryS, inTypes, outType, injectIt)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func setMacroDefaults(idx int) {
	macroDef := fmt.Sprintf("macro.%d", idx)
	viper.SetDefault(macroDef+".result", qselect.QBool)
	viper.SetDefault(macroDef+"injectIt", false)
}

func toDTypeSlice(s []string) []qselect.DType {
	d := make([]qselect.DType, 0, len(s))
	for _, t := range s {
		d = append(d, qselect.DType(t))
	}
	return d
}
