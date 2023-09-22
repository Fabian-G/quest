package config

import (
	"log"

	"github.com/Fabian-G/quest/hook"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/qsort"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/viper"
)

func buildTodoTxtRepo(v *viper.Viper, tagTypes map[string]qselect.DType) *todotxt.Repo {
	repo := todotxt.NewRepo(v.GetString(TodoFile))
	repo.DefaultHooks = hooks(v, tagTypes)
	repo.Keep = v.GetInt(KeepBackups)
	defOrder, err := qsort.CompileSortFunc(v.GetString(IdxOrder), tagTypes)
	if err != nil {
		log.Fatal(err)
	}
	repo.DefaultOrder = defOrder
	return repo
}

func hooks(v *viper.Viper, tagTypes map[string]qselect.DType) []todotxt.HookBuilder {
	hooks := make([]todotxt.HookBuilder, 0)
	if len(tagTypes) > 0 {
		hooks = append(hooks, func(l *todotxt.List) todotxt.Hook {
			return hook.NewTagExpansion(l, tagTypes)
		})
	}
	if recTag := v.GetString("recurrence.rec-tag"); recTag != "" {
		v.SetDefault("recurrence.due-tag", "due")
		v.SetDefault("recurrence.threshold-tag", "t")
		hooks = append(hooks, func(l *todotxt.List) todotxt.Hook {
			return hook.NewRecurrence(l, hook.RecurrenceTags{
				Rec:       v.GetString("recurrence.rec-tag"),
				Due:       v.GetString("recurrence.due-tag"),
				Threshold: v.GetString("recurrence.threshold-tag"),
			})
		})
	}
	return hooks
}
