package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/config"
	"github.com/Fabian-G/quest/qselect"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type prioritizeCommand struct {
	viewDef config.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newPrioritizeCommand(def config.ViewDef) *prioritizeCommand {
	cmd := prioritizeCommand{
		viewDef: def,
	}

	return &cmd
}

func (p *prioritizeCommand) command() *cobra.Command {
	var prioritizeCommand = &cobra.Command{
		Use:      "prioritize",
		Aliases:  []string{"priority", "prio"},
		Args:     cobra.MinimumNArgs(1),
		Short:    "TODO",
		Long:     `TODO `,
		Example:  "TODO",
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     p.prioritize,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	prioritizeCommand.Flags().BoolVarP(&p.all, "all", "a", false, "TODO")
	cmdutil.RegisterSelectionFlags(prioritizeCommand, &p.qql, &p.rng, &p.str)
	return prioritizeCommand
}

func (p *prioritizeCommand) prioritize(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	priority, err := todotxt.PriorityFromString(args[0])
	if err != nil {
		return fmt.Errorf("invalid priority %s: %w", args[0], err)
	}
	selector, err := cmdutil.ParseTaskSelection(p.viewDef.DefaultQuery, args[1:], p.qql, p.rng, p.str)
	if err != nil {
		return err
	}
	selection := qselect.And(notDoneFunc, selector).Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !p.all {
		confirmedSelection, err = cmdutil.ConfirmSelection(selection)
		if err != nil {
			return err
		}
	}
	if len(confirmedSelection) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no matches")
		return nil
	}

	for _, t := range confirmedSelection {
		if err := t.PrioritizeAs(priority); err != nil {
			return err
		}
	}
	cmdutil.PrintSuccessMessage(fmt.Sprintf("Prioritized as %s", priority.String()), list, confirmedSelection)
	return nil
}
