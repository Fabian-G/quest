package config

import (
	"github.com/Fabian-G/quest/hook"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/qsort"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/viper"
)

func buildTodoTxtRepo(v *viper.Viper, sortCompiler qsort.Compiler, tagTypes map[string]qselect.DType) *todotxt.Repo {
	repo := todotxt.NewRepo(v.GetString(TodoFileKey))
	repo.DefaultHooks = hooks(v, tagTypes)
	repo.Keep = v.GetInt(KeepBackupsKey)
	return repo
}

func buildDoneTxtRepo(v *viper.Viper, sortCompiler qsort.Compiler) *todotxt.Repo {
	repo := todotxt.NewRepo(v.GetString(DoneFileKey))
	repo.Keep = v.GetInt(KeepBackupsKey)
	return repo
}

func hooks(v *viper.Viper, tagTypes map[string]qselect.DType) []todotxt.HookBuilder {
	hooks := make([]todotxt.HookBuilder, 0)
	if len(tagTypes) > 0 {
		hooks = append(hooks, func(l *todotxt.List) todotxt.Hook {
			return hook.NewTagExpansion(l, v.GetBool(UnknownTagsKey), tagTypes)
		})
	}
	if recTag := v.GetString("recurrence.rec-tag"); recTag != "" {
		v.SetDefault("recurrence.due-tag", "due")
		v.SetDefault("recurrence.threshold-tag", "t")
		hooks = append(hooks, func(l *todotxt.List) todotxt.Hook {
			return hook.NewRecurrenceWithNowFunc(l, hook.RecurrenceTags{
				Rec:       v.GetString("recurrence.rec-tag"),
				Due:       v.GetString("recurrence.due-tag"),
				Threshold: v.GetString("recurrence.threshold-tag"),
			}, nowFunc(v))
		})
	}
	return hooks
}
