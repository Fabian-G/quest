package config

import (
	"log"

	"github.com/Fabian-G/quest/hook"
	"github.com/Fabian-G/quest/qsort"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/viper"
)

func buildTodoTxtRepo() *todotxt.Repo {
	repo := todotxt.NewRepo(viper.GetString(TodoFile))
	repo.DefaultHooks = hooks()
	defOrder, err := qsort.CompileSortFunc(viper.GetString(IdxOrder), TagTypes())
	if err != nil {
		log.Fatal(err)
	}
	repo.DefaultOrder = defOrder
	return repo
}

func hooks() []todotxt.HookBuilder {
	hooks := make([]todotxt.HookBuilder, 0)
	if recTag := viper.GetString("recurrence.rec-tag"); recTag != "" {
		viper.SetDefault("recurrence.due-tag", "due")
		viper.SetDefault("recurrence.threshold-tag", "t")
		hooks = append(hooks, func(l *todotxt.List) todotxt.Hook {
			return hook.NewRecurrence(l, hook.RecurrenceTags{
				Rec:       viper.GetString("recurrence.rec-tag"),
				Due:       viper.GetString("recurrence.due-tag"),
				Threshold: viper.GetString("recurrence.threshold-tag"),
			})
		})
	}
	return hooks
}
