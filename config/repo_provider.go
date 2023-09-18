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
	repo.DefaultHooks = []todotxt.HookBuilder{
		hook.NewRecurrence,
	}
	defOrder, err := qsort.CompileSortFunc(viper.GetString(DefaultSortOrder), TagTypes())
	if err != nil {
		log.Fatal(err)
	}
	repo.DefaultOrder = defOrder
	return repo
}
