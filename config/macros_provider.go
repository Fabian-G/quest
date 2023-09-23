package config

import (
	"fmt"

	"github.com/Fabian-G/quest/qselect"
	"github.com/spf13/viper"
)

type MacroDef struct {
	Name       string
	Query      string
	InTypes    []qselect.DType
	ResultType qselect.DType
	InjectIt   bool
}

func buildMacroDefs(v *viper.Viper) []MacroDef {
	macros := v.Get(Macros).([]any)
	defs := make([]MacroDef, 0, len(macros))
	for m := range macros {
		setMacroDefaults(v, m)

		macroDef := fmt.Sprintf("macro.%d", m)
		name := v.GetString(macroDef + ".name")
		queryS := v.GetString(macroDef + ".query")
		inTypes := toDTypeSlice(v.GetStringSlice(macroDef + ".args"))
		outType := qselect.DType(v.GetString(macroDef + ".result"))
		injectIt := v.GetBool(macroDef + ".injectIt")
		defs = append(defs, MacroDef{
			Name:       name,
			Query:      queryS,
			InTypes:    inTypes,
			ResultType: outType,
			InjectIt:   injectIt,
		})
	}
	return defs
}

func setMacroDefaults(v *viper.Viper, idx int) {
	macroDef := fmt.Sprintf("macro.%d", idx)
	v.SetDefault(macroDef+".result", qselect.QBool)
	v.SetDefault(macroDef+"injectIt", false)
}

func toDTypeSlice(s []string) []qselect.DType {
	d := make([]qselect.DType, 0, len(s))
	for _, t := range s {
		d = append(d, qselect.DType(t))
	}
	return d
}
