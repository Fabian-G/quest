package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/Fabian-G/quest/view"
	"github.com/spf13/cobra"
)

type prioritizeCommand struct {
	viewDef di.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newPrioritizeCommand(def di.ViewDef) *prioritizeCommand {
	cmd := prioritizeCommand{
		viewDef: def,
	}

	return &cmd
}

func (p *prioritizeCommand) command() *cobra.Command {
	var prioritizeCommand = &cobra.Command{
		Use:      "prioritize prio [selectors...]",
		Aliases:  []string{"priority", "prio"},
		Args:     cobra.MinimumNArgs(1),
		Short:    "Prioritizes all matching tasks as prio",
		Long:     "Prioritizes all matching tasks as prio\n\nYou can use None to remove priority.",
		Example:  "quest prioritize B 1,2,3",
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     p.prioritize,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	cmdutil.RegisterSelectionFlags(prioritizeCommand, &p.qql, &p.rng, &p.str, &p.all)
	return prioritizeCommand
}

func (p *prioritizeCommand) prioritize(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	priority, err := todotxt.PriorityFromString(args[0])
	if err != nil {
		return fmt.Errorf("invalid priority %s: %w", args[0], err)
	}
	selector, err := cmdutil.ParseTaskSelection(p.viewDef.Query, args[1:], p.qql, p.rng, p.str)
	if err != nil {
		return err
	}
	selection := qselect.And(notDoneFunc, selector).Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !p.all {
		confirmedSelection, err = view.NewSelection(selection).Run()
		if err != nil {
			return err
		}
	}
	if len(confirmedSelection) == 0 {
		fmt.Println("no matches")
		return nil
	}

	for _, t := range confirmedSelection {
		if err := t.PrioritizeAs(priority); err != nil {
			return err
		}
	}
	if priority != todotxt.PrioNone {
		view.NewSuccessMessage(fmt.Sprintf("Prioritized as %s", priority.String()), list, confirmedSelection).Run()
	} else {
		view.NewSuccessMessage("Cleared priority on", list, confirmedSelection).Run()
	}
	return nil
}
