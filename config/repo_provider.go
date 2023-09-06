package config

import (
	"log"

	"github.com/Fabian-G/quest/hook"
	"github.com/Fabian-G/quest/query"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/viper"
)

func buildTodoTxtRepo() *todotxt.Repo {
	repo := todotxt.NewRepo(viper.GetString(TodoFile))
	repo.DefaultHooks = []todotxt.HookBuilder{
		hook.NewRecurrence,
	}
	defOrder, err := query.SortFunc(viper.GetString(DefaultSortOrder))
	if err != nil {
		log.Fatal(err)
	}
	repo.DefaultOrder = defOrder
	return repo
}
