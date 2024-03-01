package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/hook"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/spf13/cobra"
)

type trackCommand struct {
	viewDef di.ViewDef
	qql     []string
	rng     []string
	str     []string
}

func newTrackCommand(def di.ViewDef) *trackCommand {
	cmd := trackCommand{
		viewDef: def,
	}

	return &cmd
}

func (t *trackCommand) command() *cobra.Command {
	var trackCommand = &cobra.Command{
		Use:   "track [selectors...]",
		Short: "Starts tracking the selected task with timewarrior",
		Long: `
track can be used to track the time spent on a task. 
It does so by recording the start time in a special tag called quest-tr (in minutes since epoch).
Since this alone ist not very useful it is recommended to install timewarrior.
If timewarrior is installed the quest-tr tag will trigger a hook which propagates projects, contexts and description
of the tracked task to timewarrior.
Changes that are made to a task during an active tracking are automatically reflected in timewarrior.
To stop tracking a task you can either remove the quest-tr tag from the active task or simply run "timew stop".
`,
		Example:  "quest track 3",
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     t.track,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	cmdutil.RegisterSelectionFlags(trackCommand, &t.qql, &t.rng, &t.str, nil)
	return trackCommand
}

func (t *trackCommand) track(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)

	selector, err := cmdutil.ParseTaskSelection(t.viewDef.Query, args, t.qql, t.rng, t.str)
	if err != nil {
		return err
	}

	selectedTasks := selector.Filter(list)
	if len(selectedTasks) == 0 {
		fmt.Println("no matches")
		return nil
	}
	selectedTask := selectedTasks[0]
	if len(selectedTasks) > 1 {
		t, err := selection.New("Select task to start tracking:", selectedTasks).RunPrompt()
		if err != nil {
			return fmt.Errorf("error during task selection: %w", err)
		}
		selectedTask = t
	}

	// We just set the tracking tag here. The tracking hook will do the actual work
	if err := selectedTask.SetTag(hook.TrackingTag, strconv.FormatInt(time.Now().Unix()/60, 10)); err != nil {
		return err
	}
	view.NewSuccessMessage("Started tracking", list, []*todotxt.Item{selectedTask}).Run()
	return nil
}
