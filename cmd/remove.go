package cmd

import (
	"fmt"

	"github.com/Fabian-G/quest/cmd/cmdutil"
	"github.com/Fabian-G/quest/di"
	"github.com/Fabian-G/quest/todotxt"
	"github.com/spf13/cobra"
)

type removeCommand struct {
	viewDef di.ViewDef
	qql     []string
	rng     []string
	str     []string
	all     bool
}

func newRemoveCommand(def di.ViewDef) *removeCommand {
	cmd := removeCommand{
		viewDef: def,
	}

	return &cmd
}

func (r *removeCommand) command() *cobra.Command {
	var removeCommand = &cobra.Command{
		Use:      "remove",
		Short:    "TODO",
		Long:     `TODO `,
		Example:  "TODO",
		GroupID:  "view-cmd",
		PreRunE:  cmdutil.Steps(cmdutil.LoadList),
		RunE:     r.remove,
		PostRunE: cmdutil.Steps(cmdutil.SaveList),
	}
	removeCommand.Flags().BoolVarP(&r.all, "all", "a", false, "TODO")
	cmdutil.RegisterSelectionFlags(removeCommand, &r.qql, &r.rng, &r.str)
	return removeCommand
}

func (r *removeCommand) remove(cmd *cobra.Command, args []string) error {
	list := cmd.Context().Value(cmdutil.ListKey).(*todotxt.List)
	selector, err := cmdutil.ParseTaskSelection(r.viewDef.Query, args, r.qql, r.rng, r.str)
	if err != nil {
		return err
	}
	selection := selector.Filter(list)
	var confirmedSelection []*todotxt.Item = selection
	if !r.all {
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
		if err := list.Remove(list.LineOf(t)); err != nil {
			return err
		}
	}
	cmdutil.PrintSuccessMessage("Removed", list, confirmedSelection)
	return nil
}
