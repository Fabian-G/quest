package di

import (
	"github.com/Fabian-G/quest/hook"
	"github.com/Fabian-G/quest/qsort"
	"github.com/Fabian-G/quest/todotxt"
)

func buildTodoTxtRepo(c Config, sortCompiler qsort.Compiler) *todotxt.Repo {
	repo := todotxt.NewRepo(c.TodoFile)
	repo.DefaultHooks = hooks(c)
	repo.Keep = c.KeepBackups
	return repo
}

func buildDoneTxtRepo(c Config, sortCompiler qsort.Compiler) *todotxt.Repo {
	repo := todotxt.NewRepo(c.DoneFile)
	repo.Keep = c.KeepBackups
	return repo
}

func hooks(c Config) []todotxt.Hook {
	hooks := make([]todotxt.Hook, 0)
	tagTypes := c.TagTypes()
	hooks = append(hooks, hook.NewTagExpansion(c.UnknownTags, tagTypes))

	if len(c.ClearOnDone) > 0 {
		hooks = append(hooks, hook.ClearOnDone{Clear: c.ClearOnDone})
	}
	if recTag := c.Recurrence.RecTag; recTag != "" {
		hooks = append(hooks, hook.NewRecurrence(hook.RecurrenceTags{
			Rec:       c.Recurrence.RecTag,
			Due:       c.Recurrence.DueTag,
			Threshold: c.Recurrence.ThresholdTag,
		}, hook.WithNowFunc(c.NowFunc), hook.WithPreservePriority(c.Recurrence.PreservePriority)))
	}
	return hooks
}
