package di

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

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
	if timew, err := exec.LookPath("timew"); err == nil && len(c.Tracking.Tag) > 0 {
		tracking := hook.NewTracking(c.Tracking.Tag, &timeWarrior{timew: timew})
		tracking.TrimContextPrefix = c.Tracking.TrimContextPrefix
		tracking.TrimProjectPrefix = c.Tracking.TrimProjectPrefix
		hooks = append(hooks, tracking)
	}
	return hooks
}

type timeWarrior struct {
	timew string
}

func (t *timeWarrior) ActiveTags() ([]string, error) {
	activeCmd := exec.Command(t.timew, "get", "dom.active")
	out, err := activeCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Could not determine timew active status: %w", err)
	}
	if strings.TrimSpace(string(out)) != "1" {
		return nil, hook.ErrNoActiveTracking
	}
	dataCmd := exec.Command(t.timew, "get", "dom.active.json")
	out, err = dataCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Could fetch properties of active timew interval: %w", err)
	}

	type activeJson struct {
		Tags []string `json:"tags"`
	}
	activeData := &activeJson{}
	if err := json.Unmarshal(out, activeData); err != nil {
		return nil, fmt.Errorf("Could not parse timew response: %w", err)
	}
	return activeData.Tags, nil
}

func (t *timeWarrior) SetTags(tags []string) error {
	_, err := t.ActiveTags()
	if err != nil {
		return err
	}
	args := append([]string{"retag"}, tags...)
	cmd := exec.Command(t.timew, args...)
	return cmd.Run()
}

func (t *timeWarrior) Start(tags []string) error {
	args := append([]string{"start"}, tags...)
	cmd := exec.Command(t.timew, args...)
	return cmd.Run()
}

func (t *timeWarrior) Stop() error {
	cmd := exec.Command(t.timew, "stop")
	return cmd.Run()
}
